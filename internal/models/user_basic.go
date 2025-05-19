package models

type UserBasic struct {
	Id        uint64 `gorm:"primaryKey" json:"id"`                                      // ID
	Uuid      string `gorm:"type:varchar(36);uniqueIndex;" json:"uuid"`                 // UUID
	Username  string `gorm:"type:varchar(255)" json:"username"`                         // 用户名
	Password  string `gorm:"type:varchar(36)" json:"password"`                          // 密码
	CreatedAt uint64 `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间
	UpdatedAt uint64 `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func (table *UserBasic) First() (*UserBasic, error) {
	ub := new(UserBasic)
	tx := DB.Model(table)
	if table.Id != 0 {
		tx.Where("id = ?", table.Id)
	}
	if table.Uuid != "" {
		tx.Where("uuid = ?", table.Uuid)
	}
	if table.Username != "" {
		tx.Where("username = ?", table.Username)
	}
	err := tx.First(ub).Error
	return ub, err
}

func (table *UserBasic) Create() error {
	return DB.Model(table).Create(table).Error
}
