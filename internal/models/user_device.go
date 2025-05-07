package models

type UserDevice struct {
	Id              uint64 `gorm:"primaryKey" json:"id"`                                               // ID
	UserId          uint64 `gorm:"column:user_id; index:idx_user_id" json:"user_id"`                   // 用户ID
	DeviceId        uint64 `gorm:"column:device_id" json:"device_id"`                                  // 设备ID
	DeviceCode      uint64 `gorm:"column:device_code" json:"device_code"`                              // 设备码
	ConnectPassword string `gorm:"column:connect_password; type:varchar(255)" json:"connect_password"` // 连接密码
	Remark          string `gorm:"type:varchar(255)" json:"remark"`                                    // 备注
	CreatedAt       uint64 `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"`          // 创建时间，时间戳，毫秒
	UpdatedAt       uint64 `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"`          // 更新时间，时间戳，毫秒
}

func (table *UserDevice) TableName() string {
	return "user_device"
}
