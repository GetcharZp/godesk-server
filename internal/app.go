package internal

import (
	"fmt"
	"github.com/getcharzp/godesk-serve/internal/services/channel"
	"github.com/getcharzp/godesk-serve/logger"
	pb "github.com/getcharzp/godesk-serve/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
)

func NewRpcServer() {
	listen, err := net.Listen("tcp", viper.GetString("app.port"))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterChannelServiceServer(s, &channel.Service{})

	logger.Info(fmt.Sprintf("[sys] %s start successfully, port: %s",
		viper.GetString("app.name"),
		viper.GetString("app.port")),
	)
	if err := s.Serve(listen); err != nil {
		panic(err)
	}
}
