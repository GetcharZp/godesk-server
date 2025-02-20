package models

type UserBasic struct {
	Id        uint64 `gorm:"primaryKey" json:"id"`                                      // ID
	Username  string `gorm:"type:varchar(255)" json:"username"`                         // 用户名
	Password  string `gorm:"type:varchar(36)" json:"password"`                          // 密码
	CreatedAt uint64 `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间
	UpdatedAt uint64 `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}
