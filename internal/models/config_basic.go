package models

// ConfigBasic 配置基础表
type ConfigBasic struct {
	Id        uint64 `gorm:"primaryKey;" json:"id"`                                     // ID
	Key       string `gorm:"type:varchar(255);uniqueIndex;" json:"key"`                 // 配置键
	Value     string `gorm:"type:text;" json:"value"`                                   // 配置值
	CreatedAt uint64 `gorm:"column:created_at; autoCreateTime:milli" json:"created_at"` // 创建时间，时间戳，毫秒
	UpdatedAt uint64 `gorm:"column:updated_at; autoUpdateTime:milli" json:"updated_at"` // 更新时间，时间戳，毫秒
}

func (table *ConfigBasic) TableName() string {
	return "config_basic"
}

// GetByKey 根据 key 获取配置
func (table *ConfigBasic) GetByKey(key string) (*ConfigBasic, error) {
	config := new(ConfigBasic)
	err := DB.Where("key = ?", key).First(config).Error
	return config, err
}

// GetValueByKey 根据 key 获取配置值
func (table *ConfigBasic) GetValueByKey(key string) (string, error) {
	config, err := table.GetByKey(key)
	if err != nil {
		return "", err
	}
	return config.Value, nil
}
