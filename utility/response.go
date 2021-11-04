package utility

import "github.com/gin-gonic/gin"

func ResponseData(context *gin.Context, statusCode int, msg string, data gin.H) {
	context.JSON(statusCode, gin.H{
		"code": statusCode,
		"msg":  msg,
		"data": data,
	})
}

// ResponseSuccess 成功响应
func ResponseSuccess(context *gin.Context, data gin.H) {
	ResponseData(context, 200, "ok", data)
}

// ResponseError 失败响应
func ResponseError(context *gin.Context, error string) {
	ResponseData(context, -1, error, nil)
}
