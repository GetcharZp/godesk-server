package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/getcharzp/godesk-serve/internal/services/channel"
	"github.com/getcharzp/godesk-serve/logger"
	pb "github.com/getcharzp/godesk-serve/proto"
	"google.golang.org/grpc/metadata"
)

func init() {
	// 初始化 logger
	logger.NewLogger()
}

// MockDataStreamServer 模拟 DataStreamServer
type MockDataStreamServer struct {
	ctx      context.Context
	sendChan chan *pb.ChannelRequest
}

func NewMockDataStreamServer() *MockDataStreamServer {
	return &MockDataStreamServer{
		ctx:      context.Background(),
		sendChan: make(chan *pb.ChannelRequest, 100),
	}
}

func (m *MockDataStreamServer) Send(req *pb.ChannelRequest) error {
	m.sendChan <- req
	return nil
}

func (m *MockDataStreamServer) Recv() (*pb.ChannelRequest, error) {
	// 模拟接收，阻塞等待
	<-m.ctx.Done()
	return nil, m.ctx.Err()
}

func (m *MockDataStreamServer) Context() context.Context {
	return m.ctx
}

func (m *MockDataStreamServer) SendMsg(msg interface{}) error {
	return nil
}

func (m *MockDataStreamServer) RecvMsg(msg interface{}) error {
	return nil
}

func (m *MockDataStreamServer) SetHeader(md metadata.MD) error {
	return nil
}

func (m *MockDataStreamServer) SendHeader(md metadata.MD) error {
	return nil
}

func (m *MockDataStreamServer) SetTrailer(md metadata.MD) {}

// TestScreenStreamFlow 测试完整的屏幕流流程：
// 1. 被控端注册
// 2. 控制端发送控制开始请求
// 3. 被控端响应控制开始
// 4. 被控端发送屏幕流数据
// 5. 控制端接收屏幕流数据
// 6. 控制端发送控制结束请求
// 7. 被控端响应控制结束
func TestScreenStreamFlow(t *testing.T) {
	// 创建服务
	service := &channel.Service{}

	// 模拟被控端连接
	targetUUID := "target-device-uuid"
	targetCode := uint64(123456)
	targetStream := NewMockDataStreamServer()

	// 模拟控制端连接
	controllerUUID := "controller-device-uuid"
	controllerStream := NewMockDataStreamServer()

	// 存储连接到 connMap
	connMap := channel.GetConnMap()
	connMap.Store(targetUUID, targetStream)
	connMap.Store(controllerUUID, controllerStream)

	// 1. 被控端注册
	channel.RegisterDeviceForTest(targetUUID, targetCode, "windows", "TargetPC", targetStream)
	t.Log("Step 1: 被控端注册完成")

	// 2. 控制端发送控制开始请求
	controlStartReqData, _ := json.Marshal(&pb.ControlStartedRequestData{
		TargetCode:     targetCode,
		TargetPassword: "123456",
		RequestControl: true,
		Timestamp:      time.Now().UnixMilli(),
	})
	controlStartReq := &pb.ChannelRequest{
		SendClientUuid: controllerUUID,
		Key:            "control_started_request",
		Data:           controlStartReqData,
	}

	// 处理控制开始请求
	go service.HandleRequest(controlStartReq)
	t.Log("Step 2: 控制端发送控制开始请求")

	// 3. 验证被控端收到请求
	select {
	case receivedReq := <-targetStream.sendChan:
		if receivedReq.Key != "control_started_request" {
			t.Errorf("期望收到 control_started_request，实际收到 %s", receivedReq.Key)
		}
		t.Log("Step 3: 被控端收到控制开始请求")

		// 被控端响应控制开始
		controlStartRespData, _ := json.Marshal(&pb.ControlStartedResponseData{
			Code: 0, // 成功
			Uuid: targetUUID,
		})
		controlStartResp := &pb.ChannelRequest{
			SendClientUuid:   targetUUID,
			TargetClientUuid: controllerUUID,
			Key:              "control_started_response",
			Data:             controlStartRespData,
		}
		go service.HandleRequest(controlStartResp)

	case <-time.After(2 * time.Second):
		t.Fatal("超时：被控端没有收到控制开始请求")
	}

	// 4. 验证控制端收到响应
	select {
	case receivedReq := <-controllerStream.sendChan:
		if receivedReq.Key != "control_started_response" {
			t.Errorf("期望收到 control_started_response，实际收到 %s", receivedReq.Key)
		}
		t.Log("Step 4: 控制端收到控制开始响应")

	case <-time.After(2 * time.Second):
		t.Fatal("超时：控制端没有收到控制开始响应")
	}

	// 5. 被控端发送视频帧数据（H.264格式）
	screenData, _ := json.Marshal(&pb.ScreenStreamData{
		SequenceId: 1,
		FrameData:  []byte("h264encodedframe"),
		Codec:      "h264",
		Width:      1920,
		Height:     1080,
		Timestamp:  time.Now().UnixMilli(),
		FrameType:  1, // I帧
		ExtraData:  []byte("spsppsdata"),
	})
	screenReq := &pb.ChannelRequest{
		SendClientUuid:   targetUUID,
		TargetClientUuid: controllerUUID,
		Key:              "screen_stream_data",
		Data:             screenData,
	}
	go service.HandleRequest(screenReq)
	t.Log("Step 5: 被控端发送视频帧数据")

	// 6. 验证控制端收到视频帧数据
	select {
	case receivedReq := <-controllerStream.sendChan:
		if receivedReq.Key != "screen_stream_data" {
			t.Errorf("期望收到 screen_stream_data，实际收到 %s", receivedReq.Key)
		}
		if receivedReq.SendClientUuid != targetUUID {
			t.Errorf("期望发送方是 %s，实际是 %s", targetUUID, receivedReq.SendClientUuid)
		}
		if receivedReq.TargetClientUuid != controllerUUID {
			t.Errorf("期望目标方是 %s，实际是 %s", controllerUUID, receivedReq.TargetClientUuid)
		}

		var receivedScreenData pb.ScreenStreamData
		if err := json.Unmarshal(receivedReq.Data, &receivedScreenData); err != nil {
			t.Errorf("解析视频帧数据失败: %v", err)
		}
		if receivedScreenData.SequenceId != 1 {
			t.Errorf("序列号不匹配，期望 1，实际是 %d", receivedScreenData.SequenceId)
		}
		if receivedScreenData.Codec != "h264" {
			t.Errorf("编码格式不匹配，期望 h264，实际是 %s", receivedScreenData.Codec)
		}
		if receivedScreenData.FrameType != 1 {
			t.Errorf("帧类型不匹配，期望 I帧(1)，实际是 %d", receivedScreenData.FrameType)
		}
		t.Log("Step 6: 控制端成功收到视频帧数据")

	case <-time.After(2 * time.Second):
		t.Fatal("超时：控制端没有收到屏幕流数据")
	}

	// 7. 控制端发送控制结束请求
	controlEndReqData, _ := json.Marshal(&pb.ControlEndedRequestData{
		TargetCode: targetCode,
		Timestamp:  time.Now().UnixMilli(),
	})
	controlEndReq := &pb.ChannelRequest{
		SendClientUuid: controllerUUID,
		Key:            "control_ended_request",
		Data:           controlEndReqData,
	}
	go service.HandleRequest(controlEndReq)
	t.Log("Step 7: 控制端发送控制结束请求")

	// 8. 验证被控端收到结束请求
	select {
	case receivedReq := <-targetStream.sendChan:
		if receivedReq.Key != "control_ended_request" {
			t.Errorf("期望收到 control_ended_request，实际收到 %s", receivedReq.Key)
		}
		t.Log("Step 8: 被控端收到控制结束请求")

		// 被控端响应控制结束
		controlEndRespData, _ := json.Marshal(&pb.ControlEndedResponseData{
			Code: 0, // 成功
		})
		controlEndResp := &pb.ChannelRequest{
			SendClientUuid:   targetUUID,
			TargetClientUuid: controllerUUID,
			Key:              "control_ended_response",
			Data:             controlEndRespData,
		}
		go service.HandleRequest(controlEndResp)

	case <-time.After(2 * time.Second):
		t.Fatal("超时：被控端没有收到控制结束请求")
	}

	// 9. 验证控制端收到结束响应
	select {
	case receivedReq := <-controllerStream.sendChan:
		if receivedReq.Key != "control_ended_response" {
			t.Errorf("期望收到 control_ended_response，实际收到 %s", receivedReq.Key)
		}
		t.Log("Step 9: 控制端收到控制结束响应")

	case <-time.After(2 * time.Second):
		t.Fatal("超时：控制端没有收到控制结束响应")
	}

	t.Log("完整屏幕流流程测试通过！")
}

// TestScreenStreamDataForwarding 测试视频帧数据转发
func TestScreenStreamDataForwarding(t *testing.T) {
	service := &channel.Service{}

	// 模拟连接
	targetUUID := "target-uuid"
	controllerUUID := "controller-uuid"

	targetStream := NewMockDataStreamServer()
	controllerStream := NewMockDataStreamServer()

	connMap := channel.GetConnMap()
	connMap.Store(targetUUID, targetStream)
	connMap.Store(controllerUUID, controllerStream)

	// 发送视频帧数据（H.265格式）
	screenData, _ := json.Marshal(&pb.ScreenStreamData{
		SequenceId: 100,
		FrameData:  []byte("h265encodedframe"),
		Codec:      "h265",
		Width:      2560,
		Height:     1440,
		Timestamp:  time.Now().UnixMilli(),
		FrameType:  0, // P帧
		ExtraData:  []byte("spsppsdata"),
	})
	req := &pb.ChannelRequest{
		SendClientUuid:   targetUUID,
		TargetClientUuid: controllerUUID,
		Key:              "screen_stream_data",
		Data:             screenData,
	}

	// 处理视频帧数据
	go service.HandleRequest(req)

	// 验证转发
	select {
	case forwardedReq := <-controllerStream.sendChan:
		if forwardedReq.Key != "screen_stream_data" {
			t.Errorf("Key不匹配: got %s, want screen_stream_data", forwardedReq.Key)
		}
		if forwardedReq.SendClientUuid != targetUUID {
			t.Errorf("SendClientUuid不匹配: got %s, want %s", forwardedReq.SendClientUuid, targetUUID)
		}
		if forwardedReq.TargetClientUuid != controllerUUID {
			t.Errorf("TargetClientUuid不匹配: got %s, want %s", forwardedReq.TargetClientUuid, controllerUUID)
		}

		// 验证视频帧数据
		var receivedData pb.ScreenStreamData
		if err := json.Unmarshal(forwardedReq.Data, &receivedData); err != nil {
			t.Errorf("解析视频帧数据失败: %v", err)
		}
		if receivedData.SequenceId != 100 {
			t.Errorf("序列号不匹配: got %d, want 100", receivedData.SequenceId)
		}
		if receivedData.Codec != "h265" {
			t.Errorf("编码格式不匹配: got %s, want h265", receivedData.Codec)
		}
		if receivedData.FrameType != 0 {
			t.Errorf("帧类型不匹配: got %d, want P帧(0)", receivedData.FrameType)
		}
		t.Log("视频帧数据转发成功")

	case <-time.After(2 * time.Second):
		t.Error("超时: 没有收到转发的屏幕流数据")
	}
}

// TestControlRequestForwarding 测试控制请求转发
func TestControlRequestForwarding(t *testing.T) {
	service := &channel.Service{}

	targetUUID := "target-uuid"
	controllerUUID := "controller-uuid"
	targetCode := uint64(123456)

	targetStream := NewMockDataStreamServer()
	controllerStream := NewMockDataStreamServer()

	connMap := channel.GetConnMap()
	connMap.Store(targetUUID, targetStream)
	connMap.Store(controllerUUID, controllerStream)

	// 注册被控端
	channel.RegisterDeviceForTest(targetUUID, targetCode, "windows", "TestPC", targetStream)

	// 发送控制开始请求
	controlReqData, _ := json.Marshal(&pb.ControlStartedRequestData{
		TargetCode:     targetCode,
		TargetPassword: "123456",
		RequestControl: true,
	})
	controlReq := &pb.ChannelRequest{
		SendClientUuid: controllerUUID,
		Key:            "control_started_request",
		Data:           controlReqData,
	}

	// 处理控制请求
	go service.HandleRequest(controlReq)

	// 验证被控端收到请求
	select {
	case forwardedReq := <-targetStream.sendChan:
		if forwardedReq.Key != "control_started_request" {
			t.Errorf("Key不匹配: got %s, want control_started_request", forwardedReq.Key)
		}
		t.Log("控制请求转发成功")

	case <-time.After(2 * time.Second):
		t.Error("超时: 被控端没有收到控制请求")
	}
}
