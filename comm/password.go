package comm

import "golang.org/x/crypto/bcrypt"

// Hash 把明文密码转成 bcrypt 哈希
func Hash(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Verify 校验明文密码和数据库里的哈希密码是否匹配
func Verify(hashedPassword string, rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(rawPassword))
	return err == nil
}
