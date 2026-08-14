package comm

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func NormalizeIdentity(identity string) string {
	return strings.ToUpper(strings.TrimSpace(identity))
}

// EncryptIdentity 生成不可逆的身份证号 HMAC-SHA256 摘要。
func EncryptIdentity(identity string) (string, error) {
	normalized := NormalizeIdentity(identity)
	if normalized == "" || BizConf.IdentitySecret == "" {
		return "", fmt.Errorf("identity or identity secret is empty")
	}
	mac := hmac.New(sha256.New, []byte(BizConf.IdentitySecret))
	_, _ = mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
