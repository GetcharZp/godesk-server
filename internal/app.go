package internal

import (
	"fmt"
	"net"

	"github.com/getcharzp/godesk-serve/internal/middleware"
	"github.com/getcharzp/godesk-serve/internal/services/channel"
	"github.com/getcharzp/godesk-serve/internal/services/device"
	"github.com/getcharzp/godesk-serve/internal/services/user"
	"github.com/getcharzp/godesk-serve/logger"
	pb "github.com/getcharzp/godesk-serve/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewRpcServer() {
	// 初始化设备访问 token
	if err := middleware.InitDeviceAccessToken(); err != nil {
		logger.Error("[sys] failed to load device access token", zap.Error(err))
	} else {
		logger.Info("[sys] device access token loaded successfully")
	}

	listen, err := net.Listen("tcp", viper.GetString("app.port"))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.DeviceAuth,
			middleware.UserAuth,
		),
		grpc.ChainStreamInterceptor(
			middleware.StreamDeviceAuth,
		),
	)
	pb.RegisterChannelServiceServer(s, &channel.Service{})
	pb.RegisterUserServiceServer(s, &user.Service{})
	pb.RegisterDeviceServiceServer(s, &device.Service{})

	logger.Info(fmt.Sprintf("[sys] %s start successfully, port: %s",
		viper.GetString("app.name"),
		viper.GetString("app.port")),
	)
	if err := s.Serve(listen); err != nil {
		panic(err)
	}
}
