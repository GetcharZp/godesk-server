package conf

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

func NewConfig() {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	viper.AddConfigPath(filepath.Join(p, "conf"))
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
