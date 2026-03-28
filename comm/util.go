package comm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
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

func jwtSecret() (string, error) {
	if BizConf == nil {
		return "", errors.New("biz config is not initialized")
	}
	if strings.TrimSpace(BizConf.JWTSecret) == "" {
		return "", errors.New("jwt secret is empty")
	}
	return BizConf.JWTSecret, nil
}

// GenerateToken 生成 JWT Token
func GenerateToken(openID string) (string, error) {
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}

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
	return tokenClaims.SignedString([]byte(secret))
}

// ParseToken 解析 JWT Token
func ParseToken(tokenStr string) (*JwtClaims, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}

	tokenClaims, err := jwt.ParseWithClaims(tokenStr, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
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
func AesEncrypt(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(padKey(key)))
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AesDecrypt AES 解密
func AesDecrypt(cipherStr, key string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(cipherStr)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(padKey(key)))
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func padKey(key string) string {
	if len(key) >= 32 {
		return key[:32]
	}
	return key + strings.Repeat("0", 32-len(key))
}

// IsExpired 判断是否已过报名截止时间
func IsExpired() bool {
	if BizConf == nil {
		return false
	}
	if BizConf.ExpiredDate == "" {
		return false
	}
	expiredTime, err := time.ParseInLocation(time.DateTime, BizConf.ExpiredDate, time.Local)
	if err != nil {
		return false
	}
	return time.Now().After(expiredTime)
}
