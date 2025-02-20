package models

func initTable() {
	initDeviceBasicTable()
}

func initDeviceBasicTable() {
	var seqExists bool
	if err := DB.Raw("SELECT EXISTS (SELECT 1 FROM pg_class WHERE relkind = 'S' AND relname = 'device_code_seq');").
		Scan(&seqExists).Error; err != nil {
		panic(err)
	}
	// 如果序列不存在，则创建它
	if !seqExists {
		DB.Exec("CREATE SEQUENCE device_code_seq START WITH 100000000;")
	}

	// 设置初始值
	DB.Exec("ALTER TABLE device_basic ALTER COLUMN code SET DEFAULT nextval('device_code_seq');")
}
