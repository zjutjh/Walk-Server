package middleware

import (
	"github.com/gin-gonic/gin"
)

func Auth(ctx *gin.Context)  {
	ctx.Next()
}