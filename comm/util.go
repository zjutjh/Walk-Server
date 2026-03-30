package comm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	myjwt "github.com/zjutjh/mygo/jwt"
)

// GenerateToken 生成 JWT Token
func GenerateToken(openID string) (string, error) {
	return myjwt.Pick[string]().GenerateToken(openID)
}

// GetOpenIDFromCtx 从 gin.Context 获取 OpenID（由中间件注入）
func GetOpenIDFromCtx(ctx *gin.Context) string {
	openID, err := myjwt.GetIdentity[string](ctx)
	if err != nil {
		return ""
	}
	return openID
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
