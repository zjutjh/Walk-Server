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

func ValidateBizConfig() error {
	if err := validateDateTimeConfig("biz.start_date", BizConf.StartDate); err != nil {
		return err
	}
	if err := validateDateTimeConfig("biz.expired_date", BizConf.ExpiredDate); err != nil {
		return err
	}
	if BizConf.AESSecret != "" {
		if BizConf.AESSecret == "walk_aes_secret" {
			return fmt.Errorf("biz.aes_secret must not use the example value")
		}
		if len(BizConf.AESSecret) < 16 {
			return fmt.Errorf("biz.aes_secret must be at least 16 characters")
		}
	}
	return nil
}

func validateDateTimeConfig(name, value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.ParseInLocation(time.DateTime, value, time.Local); err != nil {
		return fmt.Errorf("%s must use format %q: %w", name, time.DateTime, err)
	}
	return nil
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

func CurrentActivityDay() int {
	if BizConf.StartDate == "" {
		return 0
	}
	startTime, err := time.ParseInLocation(time.DateTime, BizConf.StartDate, time.Local)
	if err != nil {
		return 0
	}
	startDate := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	day := int(time.Since(startDate).Hours() / 24)
	if day < 0 {
		return 0
	}
	return day
}

func RouteQuotaCode(routeName string) (int, bool) {
	if BizConf.RouteQuotaCodes != nil {
		if code, ok := BizConf.RouteQuotaCodes[routeName]; ok {
			return code, true
		}
	}
	switch strings.ToLower(routeName) {
	case "zh", "zh-full", "chaohui", "chaohui-full":
		return 1, true
	case "pf-half", "pingfeng-half":
		return 2, true
	case "pf-full", "pingfeng-full":
		return 3, true
	case "mgs-half", "moganshan-half":
		return 4, true
	case "mgs", "mgs-full", "moganshan", "moganshan-full":
		return 5, true
	default:
		return 0, false
	}
}

func TeamUpperLimit(day int, routeCode int) (int, bool) {
	if BizConf.TeamUpperLimit == nil {
		return 0, false
	}
	routeLimits, ok := BizConf.TeamUpperLimit[day]
	if !ok {
		return 0, false
	}
	limit, ok := routeLimits[routeCode]
	return limit, ok
}
