package comm

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	myjwt "github.com/zjutjh/mygo/jwt"
)

func GenerateToken(userID int64) (string, error) {
	return myjwt.Pick[string]().GenerateToken(strconv.FormatInt(userID, 10))
}

func GetUserIDFromCtx(ctx *gin.Context) (int64, error) {
	id, err := myjwt.GetIdentity[string](ctx)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(id, 10, 64)
}

func NormalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

func IsValidPhone(phone string) bool {
	phone = NormalizePhone(phone)
	return len(phone) == 11 && allASCIIDigits(phone)
}

func IsValidIdentity(identity string) bool {
	identity = NormalizeIdentity(identity)
	if len(identity) != 18 || !allASCIIDigits(identity[:17]) {
		return false
	}
	last := identity[17]
	return (last >= '0' && last <= '9') || last == 'X'
}

func allASCIIDigits(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}
