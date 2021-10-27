package utility

import (
	"crypto/md5"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
	"walk-server/utility/initial"
)

// 一些结构体的定义
type jwtData struct {
	OpenID string `json:"open_id"`
	jwt.StandardClaims
}

func GenerateJWT(openID string) (string, error) {
	//设置token有效时间
	nowTime := time.Now()
	expireTime := nowTime.Add(3 * time.Hour)

	claims := jwtData{
		OpenID: fmt.Sprintf("%x", md5.Sum([]byte(openID))),
		StandardClaims: jwt.StandardClaims{
			// 过期时间
			ExpiresAt: expireTime.Unix(),
			// 指定token发行人
			Issuer: "JHWL",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//该方法内部生成签名字符串，再用于获取完整、已签名的token
	token, err := tokenClaims.SignedString([]byte(initial.Config.GetString("server.JWTSecret")))
	return token, err
}
