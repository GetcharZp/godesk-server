package util

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/getcharzp/godesk-serve/define"
	"github.com/spf13/viper"
	"github.com/up-zero/gotool/cryptoutil"
	"time"
)

// AnalyzeToken 解析token
func AnalyzeToken(token string) (*define.UserClaim, error) {
	uc := new(define.UserClaim)
	claims, err := jwt.ParseWithClaims(token, uc, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("app.jwt_key")), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return uc, errors.New("token is invalid")
	}
	return uc, err
}

// GenerateToken 生成token
func GenerateToken(uc *define.UserClaim) (string, error) {
	uc.StandardClaims = jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + 3600*24*30,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
	tokenString, err := token.SignedString([]byte(viper.GetString("app.jwt_key")))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// PasswordEncrypt 密码加密
func PasswordEncrypt(s string) string {
	p, _ := cryptoutil.Md5Iterations(s, 50)
	return p
}
