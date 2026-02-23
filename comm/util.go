package comm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JwtClaims JWT 声明
type JwtClaims struct {
	OpenID string `json:"open_id"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(openID string) (string, error) {
	claims := &JwtClaims{
		OpenID: openID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "walk-server",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenClaims.SignedString([]byte(BizConf.JWTSecret))
}

// ParseToken 解析 JWT Token
func ParseToken(tokenStr string) (*JwtClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(tokenStr, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(BizConf.JWTSecret), nil
	})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*JwtClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}

// GetOpenIDFromCtx 从 gin.Context 获取 OpenID（由中间件注入）
func GetOpenIDFromCtx(ctx *gin.Context) string {
	return ctx.GetString("open_id")
}

// AesEncrypt AES 加密
func AesEncrypt(plaintext, key string) string {
	block, err := aes.NewCipher([]byte(padKey(key)))
	if err != nil {
		return ""
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return ""
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// AesDecrypt AES 解密
func AesDecrypt(cipherStr, key string) string {
	ciphertext, err := base64.StdEncoding.DecodeString(cipherStr)
	if err != nil {
		return ""
	}
	block, err := aes.NewCipher([]byte(padKey(key)))
	if err != nil {
		return ""
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return ""
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ""
	}
	return string(plaintext)
}

func padKey(key string) string {
	if len(key) >= 32 {
		return key[:32]
	}
	return key + strings.Repeat("0", 32-len(key))
}

// IsExpired 判断是否已过报名截止时间
func IsExpired() bool {
	if BizConf.ExpiredDate == "" {
		return false
	}
	expiredTime, err := time.ParseInLocation(time.DateTime, BizConf.ExpiredDate, time.Local)
	if err != nil {
		return false
	}
	return time.Now().After(expiredTime)
}
