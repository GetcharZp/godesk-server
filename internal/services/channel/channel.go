package channel

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/getcharzp/godesk-serve/internal/models"
	"github.com/getcharzp/godesk-serve/logger"
	pb "github.com/getcharzp/godesk-serve/proto"
	"go.uber.org/zap"
)

// connMap 存储客户端连接: clientUUID -> DataStreamServer
var connMap = new(sync.Map)

// deviceManager 管理设备连接
type deviceManager struct {
	devices map[string]*deviceInfo // clientUUID -> deviceInfo
	codes   map[uint64]string      // deviceCode -> clientUUID
	mu      sync.RWMutex
}

type deviceInfo struct {
	UUID       string
	Code       uint64
	OS         string
	DeviceName string
	LastPing   time.Time
	Stream     pb.ChannelService_DataStreamServer
}

var deviceMgr = &deviceManager{
	devices: make(map[string]*deviceInfo),
	codes:   make(map[uint64]string),
}

func (dm *deviceManager) RegisterDevice(uuid string, code uint64, os, deviceName string, stream pb.ChannelService_DataStreamServer) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.devices[uuid] = &deviceInfo{
		UUID:       uuid,
		Code:       code,
		OS:         os,
		DeviceName: deviceName,
		LastPing:   time.Now(),
		Stream:     stream,
	}
	dm.codes[code] = uuid
	logger.Info("[device] registered",
		zap.String("uuid", uuid),
		zap.Uint64("code", code))
}

func (dm *deviceManager) UnregisterDevice(uuid string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if info, ok := dm.devices[uuid]; ok {
		delete(dm.codes, info.Code)
		delete(dm.devices, uuid)
		logger.Info("[device] unregistered", zap.String("uuid", uuid))
	}
}

func (dm *deviceManager) GetDeviceByCode(code uint64) (*deviceInfo, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	if uuid, ok := dm.codes[code]; ok {
		if info, ok := dm.devices[uuid]; ok {
			return info, true
		}
	}
	return nil, false
}

func (dm *deviceManager) GetDeviceByUUID(uuid string) (*deviceInfo, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	info, ok := dm.devices[uuid]
	return info, ok
}

func (dm *deviceManager) UpdateLastPing(uuid string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if info, ok := dm.devices[uuid]; ok {
		info.LastPing = time.Now()
	}
}

func (dm *deviceManager) IsOnline(code uint64) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	if uuid, ok := dm.codes[code]; ok {
		if info, ok := dm.devices[uuid]; ok {
			return time.Since(info.LastPing) < 60*time.Second
		}
	}
	return false
}

// Service channel服务
type Service struct {
	pb.UnimplementedChannelServiceServer
}

// DataStream 双向数据流处理
func (s *Service) DataStream(conn pb.ChannelService_DataStreamServer) error {
	var clientUUID string

	logger.Info("[stream] data stream started")

	for {
		req, err := conn.Recv()
		if err != nil {
			logger.Error("[stream] receive error", zap.Error(err))
			// 清理连接
			if clientUUID != "" {
				connMap.Delete(clientUUID)
				deviceMgr.UnregisterDevice(clientUUID)
			}
			return err
		}

		// 记录客户端UUID
		if req.SendClientUuid != "" && clientUUID == "" {
			clientUUID = req.SendClientUuid
			connMap.Store(clientUUID, conn)
			logger.Info("[stream] client connected", zap.String("uuid", clientUUID))
		}

		// 处理消息
		s.handleRequest(req)
	}
}

// handleRequest 处理各种消息
func (s *Service) handleRequest(req *pb.ChannelRequest) {
	logger.Info("[handle] request",
		zap.String("from", req.SendClientUuid),
		zap.String("to", req.TargetClientUuid),
		zap.String("key", req.Key))

	switch req.Key {
	case "register":
		s.handleRegister(req)
	case "heartbeat":
		s.handleHeartbeat(req)
	case "control_started_request":
		s.handleControlStartedRequest(req)
	case "control_started_response":
		s.handleControlStartedResponse(req)
	case "control_ended_request":
		s.handleControlEndedRequest(req)
	case "control_ended_response":
		s.handleControlEndedResponse(req)
	case "screen_stream_data":
		s.handleScreenStreamData(req)
	default:
		logger.Warn("[handle] unknown key", zap.String("key", req.Key))
	}
}

// handleRegister 处理设备注册
func (s *Service) handleRegister(req *pb.ChannelRequest) {
	var data pb.RegisterData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[register] unmarshal error", zap.Error(err))
		return
	}

	// 查找设备码
	device, err := (&models.DeviceBasic{Uuid: req.SendClientUuid}).First()
	if err != nil {
		logger.Error("[register] device not found", zap.String("uuid", req.SendClientUuid))
		return
	}

	// 获取连接
	if conn, ok := connMap.Load(req.SendClientUuid); ok {
		if stream, ok := conn.(pb.ChannelService_DataStreamServer); ok {
			deviceMgr.RegisterDevice(req.SendClientUuid, device.Code, data.Os, data.DeviceName, stream)
			logger.Info("[register] success",
				zap.String("uuid", req.SendClientUuid),
				zap.Uint64("code", device.Code))
		}
	}
}

// handleHeartbeat 处理心跳
func (s *Service) handleHeartbeat(req *pb.ChannelRequest) {
	deviceMgr.UpdateLastPing(req.SendClientUuid)
	logger.Info("[heartbeat] received", zap.String("uuid", req.SendClientUuid))
}

// handleControlStartedRequest 处理控制开始请求（控制端 -> 被控端）
func (s *Service) handleControlStartedRequest(req *pb.ChannelRequest) {
	var data pb.ControlStartedRequestData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[control] unmarshal error", zap.Error(err))
		return
	}

	logger.Info("[control] started request",
		zap.String("from", req.SendClientUuid),
		zap.Uint64("targetCode", data.TargetCode))

	// 查找被控端
	targetDevice, ok := deviceMgr.GetDeviceByCode(data.TargetCode)
	if !ok {
		logger.Error("[control] target device not found", zap.Uint64("code", data.TargetCode))
		// 回复错误
		s.sendResponse(req.SendClientUuid, "control_started_response", &pb.ControlStartedResponseData{
			Code: 3, // 设备不存在
			Uuid: "",
		})
		return
	}

	// 转发给被控端
	forwardReq := &pb.ChannelRequest{
		SendClientUuid:   req.SendClientUuid,
		TargetClientUuid: targetDevice.UUID,
		Key:              "control_started_request",
		Data:             req.Data,
	}

	if err := s.sendTo(forwardReq, targetDevice.UUID); err != nil {
		logger.Error("[control] forward error", zap.Error(err))
		s.sendResponse(req.SendClientUuid, "control_started_response", &pb.ControlStartedResponseData{
			Code: 4, // 设备离线
			Uuid: "",
		})
	}
}

// handleControlStartedResponse 处理控制开始响应（被控端 -> 控制端）
func (s *Service) handleControlStartedResponse(req *pb.ChannelRequest) {
	var data pb.ControlStartedResponseData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[control] unmarshal error", zap.Error(err))
		return
	}

	// 从被控端获取设备码
	if deviceInfo, ok := deviceMgr.GetDeviceByUUID(req.SendClientUuid); ok {
		data.TargetCode = deviceInfo.Code
	}

	logger.Info("[control] started response",
		zap.String("from", req.SendClientUuid),
		zap.String("to", req.TargetClientUuid),
		zap.Int32("code", data.Code),
		zap.Uint64("targetCode", data.TargetCode))

	// 重新序列化数据（包含设备码）
	updatedData, _ := json.Marshal(&data)
	forwardReq := &pb.ChannelRequest{
		SendClientUuid:   req.SendClientUuid,
		TargetClientUuid: req.TargetClientUuid,
		Key:              req.Key,
		Data:             updatedData,
	}

	// 转发给控制端
	if err := s.sendTo(forwardReq, req.TargetClientUuid); err != nil {
		logger.Error("[control] forward response error", zap.Error(err))
	}
}

// handleControlEndedRequest 处理控制结束请求
func (s *Service) handleControlEndedRequest(req *pb.ChannelRequest) {
	var data pb.ControlEndedRequestData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[control] unmarshal error", zap.Error(err))
		return
	}

	logger.Info("[control] ended request",
		zap.String("from", req.SendClientUuid),
		zap.Uint64("targetCode", data.TargetCode))

	// 查找被控端
	targetDevice, ok := deviceMgr.GetDeviceByCode(data.TargetCode)
	if !ok {
		logger.Error("[control] target device not found", zap.Uint64("code", data.TargetCode))
		return
	}

	// 转发给被控端
	forwardReq := &pb.ChannelRequest{
		SendClientUuid:   req.SendClientUuid,
		TargetClientUuid: targetDevice.UUID,
		Key:              "control_ended_request",
		Data:             req.Data,
	}

	if err := s.sendTo(forwardReq, targetDevice.UUID); err != nil {
		logger.Error("[control] forward error", zap.Error(err))
	}
}

// handleControlEndedResponse 处理控制结束响应
func (s *Service) handleControlEndedResponse(req *pb.ChannelRequest) {
	var data pb.ControlEndedResponseData
	if err := json.Unmarshal(req.Data, &data); err != nil {
		logger.Error("[control] unmarshal error", zap.Error(err))
		return
	}

	logger.Info("[control] ended response",
		zap.String("from", req.SendClientUuid),
		zap.String("to", req.TargetClientUuid),
		zap.Int32("code", data.Code))

	// 转发给控制端
	if err := s.sendTo(req, req.TargetClientUuid); err != nil {
		logger.Error("[control] forward response error", zap.Error(err))
	}
}

// handleScreenStreamData 处理屏幕流数据（被控端 -> 控制端）
func (s *Service) handleScreenStreamData(req *pb.ChannelRequest) {
	logger.Info("[screen] stream data",
		zap.String("from", req.SendClientUuid),
		zap.String("to", req.TargetClientUuid),
		zap.Int("size", len(req.Data)))

	if req.TargetClientUuid == "" {
		logger.Error("[screen] target_client_uuid is empty")
		return
	}

	// 直接转发给控制端
	if err := s.sendTo(req, req.TargetClientUuid); err != nil {
		logger.Error("[screen] forward error", zap.Error(err))
	}
}

// sendTo 发送消息到指定客户端
func (s *Service) sendTo(req *pb.ChannelRequest, targetUUID string) error {
	if conn, ok := connMap.Load(targetUUID); ok {
		if stream, ok := conn.(pb.ChannelService_DataStreamServer); ok {
			return stream.Send(req)
		}
	}
	return errors.New("client not found: " + targetUUID)
}

// sendResponse 发送响应
func (s *Service) sendResponse(targetUUID, key string, data interface{}) {
	dataBytes, _ := json.Marshal(data)
	req := &pb.ChannelRequest{
		SendClientUuid:   "server",
		TargetClientUuid: targetUUID,
		Key:              key,
		Data:             dataBytes,
	}
	s.sendTo(req, targetUUID)
}

// IsDeviceOnline 检查设备是否在线（供其他包使用）
func IsDeviceOnline(code uint64) bool {
	return deviceMgr.IsOnline(code)
}

// ==================== 测试支持 ====================

// GetConnMap 获取连接映射（仅用于测试）
func GetConnMap() *sync.Map {
	return connMap
}

// GetDeviceManager 获取设备管理器（仅用于测试）
func GetDeviceManager() *deviceManager {
	return deviceMgr
}

// HandleRequest 处理请求（仅用于测试）
func (s *Service) HandleRequest(req *pb.ChannelRequest) {
	s.handleRequest(req)
}

// RegisterDeviceForTest 注册设备用于测试
func RegisterDeviceForTest(uuid string, code uint64, os, deviceName string, stream pb.ChannelService_DataStreamServer) {
	deviceMgr.RegisterDevice(uuid, code, os, deviceName, stream)
}
