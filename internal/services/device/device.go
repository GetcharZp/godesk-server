package device

import (
	"context"
	"github.com/getcharzp/godesk-serve/define"
	"github.com/getcharzp/godesk-serve/internal/models"
	"github.com/getcharzp/godesk-serve/internal/resp"
	"github.com/getcharzp/godesk-serve/internal/util"
	"github.com/getcharzp/godesk-serve/logger"
	godesk "github.com/getcharzp/godesk-serve/proto"
	"github.com/up-zero/gotool/convertutil"
	"github.com/up-zero/gotool/idutil"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetDeviceInfo 获取设备信息
func (s *Service) GetDeviceInfo(ctx context.Context, in *godesk.DeviceInfoRequest) (*godesk.DeviceInfoResponse, error) {
	db, err := (&models.DeviceBasic{Uuid: in.Uuid}).First()
	if err != nil {
		logger.Error("[db] get device basic error.", zap.Error(err))
		return nil, err
	}
	reply := new(godesk.DeviceInfoResponse)
	if err := convertutil.CopyProperties(db, reply); err != nil {
		logger.Error("[sys] copy properties error.", zap.Error(err))
		return nil, err
	}

	return reply, nil
}

// CreateDevice 创建设备
func (s *Service) CreateDevice(ctx context.Context, in *godesk.CreateDeviceRequest) (*godesk.DeviceInfoResponse, error) {
	db := &models.DeviceBasic{
		Uuid:     idutil.UUIDGenerate(),
		Os:       in.Os,
		RemoteIp: "",
	}
	if err := models.DB.Create(db).Error; err != nil {
		logger.Error("[db] create device basic error.", zap.Error(err))
		return nil, err
	}
	reply := new(godesk.DeviceInfoResponse)
	if err := convertutil.CopyProperties(db, reply); err != nil {
		logger.Error("[sys] copy properties error.", zap.Error(err))
		return nil, err
	}

	return reply, nil
}

// GetDeviceList 获取设备列表
func (s *Service) GetDeviceList(ctx context.Context, in *godesk.DeviceListRequest) (*godesk.DeviceListResponse, error) {
	util.InitBaseRequest(in)
	uc := ctx.Value("user_claims").(*define.UserClaim)
	reply, err := (&models.UserDevice{UserUuid: uc.Uuid}).List(in)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return reply, nil
}

// AddDevice 添加设备
func (s *Service) AddDevice(ctx context.Context, in *godesk.AddDeviceRequest) (*godesk.EmptyResponse, error) {
	uc := ctx.Value("user_claims").(*define.UserClaim)
	_, err := (&models.DeviceBasic{Code: in.Code}).First()
	if err != nil {
		logger.Error("[db] get device basic error.", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, resp.MsgDeviceCodeNotExist)
	}

	ud := &models.UserDevice{
		Uuid:           idutil.UUIDGenerate(),
		UserUuid:       uc.Uuid,
		DeviceCode:     in.Code,
		DevicePassword: in.Password,
		Remark:         in.Remark,
	}
	cnt, err := ud.CountForSave()
	if err != nil {
		logger.Error("[db] count user device error.", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if cnt > 0 {
		return nil, status.Errorf(codes.AlreadyExists, resp.MsgDeviceCodeExist)
	}
	if err = models.DB.Create(ud).Error; err != nil {
		logger.Error("[db] create user device error.", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return nil, nil
}

// EditDevice 修改设备
func (s *Service) EditDevice(ctx context.Context, in *godesk.EditDeviceRequest) (*godesk.EmptyResponse, error) {
	uc := ctx.Value("user_claims").(*define.UserClaim)
	cnt, err := (&models.UserDevice{
		Uuid:       in.Uuid,
		UserUuid:   uc.Uuid,
		DeviceCode: in.Code,
	}).CountForSave()
	if err != nil {
		logger.Error("[db] count user device error.", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if cnt > 0 {
		return nil, status.Errorf(codes.AlreadyExists, resp.MsgDeviceCodeExist)
	}
	if err := models.DB.Model(&models.UserDevice{}).Where("uuid = ? AND user_uuid = ?", in.Uuid, uc.Uuid).Updates(map[string]any{
		"device_code":     in.Code,
		"device_password": in.Password,
		"remark":          in.Remark,
	}).Error; err != nil {
		logger.Error("[db] update user device error.", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return nil, nil
}

// DeleteDevice 删除设备
func (s *Service) DeleteDevice(ctx context.Context, in *godesk.DeleteDeviceRequest) (*godesk.EmptyResponse, error) {
	uc := ctx.Value("user_claims").(*define.UserClaim)
	if err := models.DB.Delete(&models.UserDevice{}, "uuid = ? AND user_uuid = ?", in.Uuid, uc.Uuid).Error; err != nil {
		logger.Error("[db] delete user device error.", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return nil, nil
}
