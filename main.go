package main

import (
	"github.com/getcharzp/godesk-serve/conf"
	"github.com/getcharzp/godesk-serve/internal"
	"github.com/getcharzp/godesk-serve/logger"
)

func main() {
	// 初始化配置
	conf.NewConfig()
	// 初始化日志
	logger.NewLogger()

	// 启动服务
	internal.NewRpcServer()
}
