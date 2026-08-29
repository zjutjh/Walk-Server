package comm

import (
	"strconv"

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
	return strconv.ParseInt(id 10, 64)
}
