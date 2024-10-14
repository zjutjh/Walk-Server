/*
 * Copyright (c) 2021 IInfo.
 */

package initial

import (
	"fmt"
	"io"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RouterInit modify & initial gin router
func RouterInit() *gin.Engine {
	// 创建 gin 实例
	router := gin.New()

	// 打开日志文件，并设置为追加模式，防止覆盖
	logFile, err := os.OpenFile("gin.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	// 将日志输出到终端和文件
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)

	// 自定义日志格式，符合 GoAccess 格式
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %d\n",
			param.ClientIP,     // 客户端 IP
			param.Request.Host, // 请求的 Host（可选）
			param.TimeStamp.Format("02/Jan/2006:15:04:05"), // 时间戳
			param.Method,                 // HTTP 方法
			param.Path,                   // 请求路径
			param.Request.Proto,          // 协议版本
			param.StatusCode,             // HTTP 状态码
			param.BodySize,               // 响应体大小
			param.Request.UserAgent(),    // 用户代理
			param.ErrorMessage,           // 错误信息
			param.Latency.Microseconds(), // 响应时间 (RT) 单位：微秒
		)
	}))

	// 使用 gin 的恢复中间件来捕获 panic
	router.Use(gin.Recovery())

	// 配置 CORS
	config := cors.DefaultConfig()
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	config.AllowAllOrigins = true
	router.Use(cors.New(config))

	return router
}
