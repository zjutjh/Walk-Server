package middleware

import (
	"github.com/gin-gonic/gin"
	"walk-server/utility"
)

func Intercept(context *gin.Context) {
	utility.ResponseError(context, "该操作不允许")
	context.Abort()
	return
}
