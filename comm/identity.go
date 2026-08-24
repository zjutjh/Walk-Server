package comm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

func NormalizeIdentity(identity string) string {
	return strings.ToUpper(strings.TrimSpace(identity))
}

// EncryptIdentity 使用确定性 AES-GCM 加密身份证号。
// 相同身份证号会生成相同密文，便于数据库查重和校验。
func EncryptIdentity(identity string) (string, error) {
	normalized := NormalizeIdentity(identity)
	if normalized == "" || BizConf.IdentitySecret == "" {
		return "", fmt.Errorf("identity or identity secret is empty")
	}
	gcm, key, err := identityGCM()
	if err != nil {
		return "", err
	}
	nonce := identityNonce(key, normalized, gcm.NonceSize())
	ciphertext := gcm.Seal(nil, nonce, []byte(normalized), nil)
	payload := append(append(make([]byte, 0, len(nonce)+len(ciphertext)), nonce...), ciphertext...)
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func identityGCM() (cipher.AEAD, []byte, error) {
	keySum := sha256.Sum256([]byte(BizConf.IdentitySecret))
	key := keySum[:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("create identity cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("create identity GCM: %w", err)
	}
	return gcm, key, nil
}

func identityNonce(key []byte, identity string, size int) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte("identity-nonce:" + identity))
	return mac.Sum(nil)[:size]
}
