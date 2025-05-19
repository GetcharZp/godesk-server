package models

import (
	"github.com/getcharzp/godesk-serve/internal/util"
	"github.com/getcharzp/godesk-serve/logger"
	godesk "github.com/getcharzp/godesk-serve/proto"
	"go.uber.org/zap"
)

type UserDevice struct {
	Id             uint64 `gorm:"primaryKey" json:"id"`                                             // ID
	Uuid           string `gorm:"type:varchar(36);uniqueIndex;" json:"uuid"`                        // 记录唯一标识
	UserUuid       string `gorm:"column:user_uuid; index:idx_user_uuid" json:"user_uuid"`           // 用户唯一标识
	DeviceCode     uint64 `gorm:"column:device_code" json:"device_code"`                            // 设备码
	DevicePassword string `gorm:"column:device_password; type:varchar(255)" json:"device_password"` // 连接密码
	Remark         string `gorm:"type:varchar(255)" json:"remark"`                                  // 备注
	CreatedAt      uint64 `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"`        // 创建时间，时间戳，毫秒
	UpdatedAt      uint64 `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"`        // 更新时间，时间戳，毫秒
}

func (table *UserDevice) TableName() string {
	return "user_device"
}

// CountForSave 用户设备数量（新增 || 修改）
func (table *UserDevice) CountForSave() (int64, error) {
	var cnt int64
	tx := DB.Model(table).Where("user_uuid = ? AND device_code = ?", table.UserUuid, table.DeviceCode)
	if table.Uuid != "" {
		tx.Where("uuid != ?", table.Uuid)
	}
	err := tx.Count(&cnt).Error
	return cnt, err
}

// List 用户设备列表
func (table *UserDevice) List(in *godesk.DeviceListRequest) (*godesk.DeviceListResponse, error) {
	var list []*godesk.DeviceListItem
	var cnt int64
	var pager = util.NewPager(in.Base)
	tx := DB.Model(table).Select("user_device.uuid, "+
		"user_device.device_code code, "+
		"user_device.remark, "+
		"device_basic.os, "+
		"user_device.device_password password").
		Joins("LEFT JOIN device_basic ON user_device.device_code = device_basic.code").
		Where("user_device.user_uuid = ?", table.UserUuid)
	if in.Base.Keyword != "" {
		tx.Where("user_device.device_code LIKE ?", "%"+in.Base.Keyword+"%")
	}
	err := tx.Count(&cnt).Offset(pager.Offset()).Limit(pager.Limit()).Find(&list).Error
	if err != nil {
		logger.Error("[db] get user device list error.", zap.Error(err))
		return nil, err
	}
	return &godesk.DeviceListResponse{
		Count: cnt,
		List:  list,
	}, nil
}
