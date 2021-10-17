package middleware

import "github.com/gin-gonic/gin"

//TimeValidity Require implement ... Check if in open time
func TimeValidity(ctx *gin.Context) {
	ctx.Next()
}
