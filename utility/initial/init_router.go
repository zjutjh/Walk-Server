/*
 * Copyright (c) 2021 IInfo.
 */

package initial

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//RouterInit modify & initial gin router
func RouterInit() *gin.Engine {
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

	// 给 API 添加上跨域功能
	config := cors.DefaultConfig()
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	return router
}
