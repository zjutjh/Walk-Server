package utility

import (
	"strings"
	"time"
	"walk-server/global"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// JwtData 一些结构体的定义
type JwtData struct {
	OpenID string `json:"open_id"`
	jwt.StandardClaims
}

// GenerateStandardJwt 根据数据生成带有 standard claims 的 jwt token
func GenerateStandardJwt(jwtData *JwtData) (string, error) {
	claims := jwtData
	claims.StandardClaims = jwt.StandardClaims{
		// 过期时间
		ExpiresAt: time.Now().Add(168 * time.Hour).Unix(), // 设置 7 天后过期
		// 指定token发行人
		Issuer: "JHWL",
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//该方法内部生成签名字符串，再用于获取完整、已签名的token
	token, err := tokenClaims.SignedString([]byte(global.Config.GetString("server.JWTSecret")))
	return token, err
}

// ParseToken 根据传入的token值获取到Claims对象信息，（进而获取其中的用户名和密码）
func ParseToken(token string) (*JwtData, error) {
	jwtSecret := []byte(global.Config.GetString("server.JWTSecret"))
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

// UrlToken 用来生成能在 url 中传输的
func UrlToken(jwtData *JwtData) (string, error) {
	jwtToken, err := GenerateStandardJwt(jwtData)

	// 将 jwtToken 中的 . 符号替换成能在 http 中传输的 !
	jwtToken = strings.Replace(jwtToken, ".", "!", 2)

	return jwtToken, err
}

// GetJwtData 从控制器上下文中获取 jwt 数据
func GetJwtData(context *gin.Context) *JwtData {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, err := ParseToken(jwtToken)
	if err != nil {
		return nil
	}
	return jwtData
}
