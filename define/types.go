package define

import "github.com/dgrijalva/jwt-go"

type UserClaim struct {
	Uuid     string // 用户唯一标识
	Username string // 登录名
	jwt.StandardClaims
}
