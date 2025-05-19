package user

import (
	"context"
	"errors"
	"github.com/getcharzp/godesk-serve/define"
	"github.com/getcharzp/godesk-serve/internal/models"
	"github.com/getcharzp/godesk-serve/internal/resp"
	"github.com/getcharzp/godesk-serve/internal/util"
	"github.com/getcharzp/godesk-serve/logger"
	godesk "github.com/getcharzp/godesk-serve/proto"
	"github.com/up-zero/gotool/idutil"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// GetUserInfo 获取用户信息
func (s *Service) GetUserInfo(ctx context.Context, in *godesk.EmptyRequest) (*godesk.UserInfoResponse, error) {
	uc := ctx.Value("user_claims").(*define.UserClaim)
	token, err := util.GenerateToken(&define.UserClaim{
		Uuid:     uc.Uuid,
		Username: uc.Username,
	})
	if err != nil {
		logger.Error("[sys] generate token error.", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	reply := &godesk.UserInfoResponse{
		Uuid:     uc.Uuid,
		Username: uc.Username,
		Token:    token,
	}
	return reply, nil
}

// UserRegister 用户注册
func (s *Service) UserRegister(ctx context.Context, in *godesk.UserRegisterRequest) (*godesk.UserInfoResponse, error) {
	ub := &models.UserBasic{
		Username: in.Username,
		Password: util.PasswordEncrypt(in.Password),
	}
	_, err := ub.First()
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.AlreadyExists, resp.MsgUserExist)
	}
	ub.Uuid = idutil.UUIDGenerate()
	if err = ub.Create(); err != nil {
		logger.Error("[db] create user basic error.", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	token, err := util.GenerateToken(&define.UserClaim{
		Uuid:     ub.Uuid,
		Username: ub.Username,
	})
	if err != nil {
		logger.Error("[sys] generate token error.", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	reply := &godesk.UserInfoResponse{
		Uuid:     ub.Uuid,
		Username: ub.Username,
		Token:    token,
	}
	return reply, nil
}

// UserLogin 用户登录
func (s *Service) UserLogin(ctx context.Context, in *godesk.UserLoginRequest) (*godesk.UserInfoResponse, error) {
	ub := &models.UserBasic{
		Username: in.Username,
	}
	ub, err := ub.First()
	if err != nil {
		logger.Error("[db] get user basic error.", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, resp.MsgUserNotExist)
	}
	if ub.Password != util.PasswordEncrypt(in.Password) {
		return nil, status.Errorf(codes.InvalidArgument, resp.MsgUserPassword)
	}
	token, err := util.GenerateToken(&define.UserClaim{
		Uuid:     ub.Uuid,
		Username: ub.Username,
	})
	reply := &godesk.UserInfoResponse{
		Uuid:     ub.Uuid,
		Username: ub.Username,
		Token:    token,
	}
	return reply, nil
}
