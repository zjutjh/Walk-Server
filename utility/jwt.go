package utility

import (
	"crypto/md5"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
	"walk-server/utility/initial"
)

// JwtData 一些结构体的定义
type JwtData struct {
	OpenID   string `json:"open_id"`
	Identity string `json:"identity"`
	jwt.StandardClaims
}

func GenerateJWT(openID string) (string, error) {
	//设置token有效时间
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour) // 1 天后过期

	claims := JwtData{
		OpenID:   fmt.Sprintf("%x", md5.Sum([]byte(openID))),
		Identity: "not-join", // 刚注册完初始化身份为 not-join
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

// ParseToken 根据传入的token值获取到Claims对象信息，（进而获取其中的用户名和密码）
func ParseToken(token string) (*JwtData, error) {
	jwtSecret := []byte(initial.Config.GetString("server.JWTSecret"))
	//用于解析鉴权的声明，方法内部主要是具体的解码和校验的过程，最终返回*Token
	tokenClaims, err := jwt.ParseWithClaims(token, &JwtData{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		// 从tokenClaims中获取到Claims对象，并使用断言，将该对象转换为我们自己定义的Claims
		// 要传入指针，项目中结构体都是用指针传递，节省空间。
		if claims, ok := tokenClaims.Claims.(*JwtData); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
