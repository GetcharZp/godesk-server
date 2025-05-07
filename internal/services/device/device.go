package device

import (
	"context"
	"github.com/getcharzp/godesk-serve/internal/models"
	"github.com/getcharzp/godesk-serve/logger"
	godesk "github.com/getcharzp/godesk-serve/proto"
	"github.com/up-zero/gotool/convertutil"
	"github.com/up-zero/gotool/idutil"
	"go.uber.org/zap"
)

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
