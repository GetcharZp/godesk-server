package middleware

import (
	"context"
	"github.com/getcharzp/godesk-serve/internal/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = map[string]struct{}{
	"/godesk.UserService/UserLogin":    {},
	"/godesk.UserService/UserRegister": {},
}

// UserAuth 用户鉴权
func UserAuth(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if _, ok := publicMethods[info.FullMethod]; ok {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}
	tokens := md["authorization"]
	if len(tokens) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}
	uc, err := util.AnalyzeToken(tokens[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	ctx = context.WithValue(ctx, "user_claims", uc)
	return handler(ctx, req)
}
