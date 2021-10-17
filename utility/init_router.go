/*
 * Copyright (c) 2021 IInfo.
 */

package utility

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

//InitRouter modify & init gin router
func InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// log format
		return fmt.Sprintf("%s |%s %d %s| %s |%s %s %s %s | %s | size: %d | %s | %s\n",
			param.TimeStamp.Format(time.RFC1123),
			param.StatusCodeColor(),
			param.StatusCode,
			param.ResetColor(),
			param.ClientIP,
			param.MethodColor(),
			param.Method,
			param.ResetColor(),
			param.Path,
			param.Latency,
			param.BodySize,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())

	return router
}
