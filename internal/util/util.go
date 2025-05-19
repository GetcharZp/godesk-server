package util

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/getcharzp/godesk-serve/define"
	godesk "github.com/getcharzp/godesk-serve/proto"
	"github.com/spf13/viper"
	"github.com/up-zero/gotool/cryptoutil"
	"reflect"
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

// InitBaseRequest  初始化BaseRequest
func InitBaseRequest(obj any) {
	v := reflect.ValueOf(obj).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if field.Type == reflect.TypeOf(&godesk.BaseRequest{}) {
			if value.IsNil() {
				value.Set(reflect.ValueOf(&godesk.BaseRequest{
					Page: define.DefaultPage,
					Size: define.DefaultSize,
				}))
			} else {
				br := value.Interface().(*godesk.BaseRequest)
				if br.Page == 0 {
					br.Page = define.DefaultPage
				}
				if br.Size == 0 {
					br.Size = define.DefaultSize
				}
			}
		}
	}
}
