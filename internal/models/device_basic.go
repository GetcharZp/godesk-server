package models

type DeviceBasic struct {
	Id        uint64 `gorm:"primaryKey;" json:"id"`                                     // ID
	Uuid      string `gorm:"type:varchar(36);uniqueIndex;" json:"uuid"`                 // UUID
	Code      uint64 `gorm:"uniqueIndex;autoIncrement;" json:"code"`                    // 设备码
	Os        string `gorm:"type:varchar(255)" json:"os"`                               // 操作系统, win, mac, linux
	RemoteIp  string `gorm:"column:remote_ip; type:varchar(255)" json:"remote_ip"`      // 远程IP
	CreatedAt uint64 `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt uint64 `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间，时间戳，毫秒
}

func (table *DeviceBasic) TableName() string {
	return "device_basic"
}

func (table *DeviceBasic) First() (*DeviceBasic, error) {
	db := new(DeviceBasic)
	tx := DB.Model(table)
	if table.Uuid != "" {
		tx = tx.Where("uuid = ?", table.Uuid)
	}
	if table.Code != 0 {
		tx = tx.Where("code = ?", table.Code)
	}
	err := tx.First(db).Error
	return db, err
}
