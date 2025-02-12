package channel

import (
	"errors"
	"github.com/getcharzp/godesk-serve/logger"
	pb "github.com/getcharzp/godesk-serve/proto"
	"go.uber.org/zap"
	"sync"
)

var connMap = new(sync.Map)

func (in *Service) DataStream(conn pb.ChannelService_DataStreamServer) error {
	for {
		req, err := conn.Recv()
		if err != nil {
			logger.Error("[sys] receive data error.", zap.Error(err))
			return err
		}
		connMap.Store(req.ClientUuid, conn)
		in.ReceiveDataHandle(req)
	}
}

func (in *Service) ReceiveDataHandle(req *pb.ChannelRequest) {
	logger.Info("[sys] receive data.", zap.Any("data", req))
	// todo: 接受数据后的处理逻辑
}

// SendMessage 单发消息
func (in *Service) SendMessage(req *pb.ChannelRequest) error {
	value, ok := connMap.Load(req.ClientUuid)
	if !ok {
		logger.Info("[sys] client not fount.", zap.String("ClientUUID", req.GetClientUuid()))
		return errors.New("client not found")
	}
	conn, ok := value.(pb.ChannelService_DataStreamServer)
	if err := conn.Send(req); err != nil {
		logger.Error("[sys] send message error.", zap.Error(err))
		return err
	}

	return nil
}
