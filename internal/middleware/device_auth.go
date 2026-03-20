package middleware

import (
	"context"
	"github.com/getcharzp/godesk-serve/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sync"
)

// deviceAccessToken 存储从数据库加载的 access token
var deviceAccessToken string
var tokenOnce sync.Once

// InitDeviceAccessToken 从数据库加载 access token
// 在应用启动时调用
func InitDeviceAccessToken() error {
	var err error
	tokenOnce.Do(func() {
		config := &models.ConfigBasic{}
		deviceAccessToken, err = config.GetValueByKey("device_access_token")
	})
	return err
}

// GetDeviceAccessToken 获取当前配置的 access token
func GetDeviceAccessToken() string {
	return deviceAccessToken
}

// SetDeviceAccessToken 设置 access token（用于动态更新）
func SetDeviceAccessToken(token string) {
	deviceAccessToken = token
}

// devicePublicMethods 不需要 token 验证的公开方法
var devicePublicMethods = map[string]struct{}{
	// 如果有不需要 token 的方法，在这里添加
	// "/godesk.ChannelService/DataStream": {}, // 通道服务可能需要特殊处理
}

// DeviceAuth 设备鉴权中间件
// 验证请求头中是否携带了正确的 access token
func DeviceAuth(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// 检查是否是公开方法
	if _, ok := devicePublicMethods[info.FullMethod]; ok {
		return handler(ctx, req)
	}

	// 从 metadata 中获取 access token
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// 获取 accesstoken（支持小写和大写）
	tokens := md["accesstoken"]
	if len(tokens) == 0 {
		tokens = md["accessToken"]
	}
	if len(tokens) == 0 {
		tokens = md["AccessToken"]
	}
	if len(tokens) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing access token")
	}

	// 验证 token
	if deviceAccessToken == "" {
		return nil, status.Errorf(codes.Internal, "device access token not configured")
	}

	if tokens[0] != deviceAccessToken {
		return nil, status.Errorf(codes.Unauthenticated, "invalid access token")
	}

	return handler(ctx, req)
}

// StreamDeviceAuth 流式设备的鉴权中间件
func StreamDeviceAuth(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 检查是否是公开方法
	if _, ok := devicePublicMethods[info.FullMethod]; ok {
		return handler(srv, ss)
	}

	// 从 metadata 中获取 access token
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// 获取 accesstoken（支持小写和大写）
	tokens := md["accesstoken"]
	if len(tokens) == 0 {
		tokens = md["accessToken"]
	}
	if len(tokens) == 0 {
		tokens = md["AccessToken"]
	}
	if len(tokens) == 0 {
		return status.Errorf(codes.Unauthenticated, "missing access token")
	}

	// 验证 token
	if deviceAccessToken == "" {
		return status.Errorf(codes.Internal, "device access token not configured")
	}

	if tokens[0] != deviceAccessToken {
		return status.Errorf(codes.Unauthenticated, "invalid access token")
	}

	return handler(srv, ss)
}
