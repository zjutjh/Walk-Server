package comm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	myjwt "github.com/zjutjh/mygo/jwt"
)

// GenerateToken 生成 JWT Token
func GenerateToken(userID int64) (string, error) {
	return myjwt.Pick[string]().GenerateToken(strconv.FormatInt(userID, 10))
}

// GetUserIDFromCtx 从 gin.Context 获取用户 ID（由 JWT 中间件注入）。
func GetUserIDFromCtx(ctx *gin.Context) (int64, error) {
	identity, err := myjwt.GetIdentity[string](ctx)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(identity, 10, 64)
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

func NormalizeIdentity(identity string) string {
	return strings.ToUpper(strings.TrimSpace(identity))
}

// EncryptIdentity 将身份证号标准化后生成不可逆的 HMAC-SHA256 摘要。
// 数据库只保存该摘要，不保存身份证号明文。
func EncryptIdentity(identity string) (string, error) {
	normalized := NormalizeIdentity(identity)
	if normalized == "" || BizConf.IdentitySecret == "" {
		return "", fmt.Errorf("identity or identity secret is empty")
	}
	mac := hmac.New(sha256.New, []byte(BizConf.IdentitySecret))
	_, _ = mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil)), nil
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
	if len(BizConf.IdentitySecret) < 32 {
		return fmt.Errorf("biz.identity_secret must be at least 32 characters")
	}
	if BizConf.TeamTotalLimit <= 0 {
		return fmt.Errorf("biz.team_total_limit must be greater than 0")
	}
	if len(BizConf.DailyTeamLimits) == 0 {
		return fmt.Errorf("biz.daily_team_limits must not be empty")
	}
	for day, limit := range BizConf.DailyTeamLimits {
		if limit < 0 {
			return fmt.Errorf("biz.daily_team_limits[%d] must not be negative", day)
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

// IsInRegisterTime 判断当前时间是否处于队伍提交时间段。
func IsInRegisterTime() bool {
	now := time.Now()
	if BizConf.StartDate != "" {
		start, err := time.ParseInLocation(time.DateTime, BizConf.StartDate, time.Local)
		if err != nil || now.Before(start) {
			return false
		}
	}
	if BizConf.ExpiredDate != "" {
		expired, err := time.ParseInLocation(time.DateTime, BizConf.ExpiredDate, time.Local)
		if err != nil || now.After(expired) {
			return false
		}
	}
	return true
}

func DailyTeamLimit(day int) (int, bool) {
	if day < 0 || day >= len(BizConf.DailyTeamLimits) {
		return 0, false
	}
	return BizConf.DailyTeamLimits[day], true
}
