package main

import (
	"walk-server/global"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func main() {
	initial.ConfigInit()   // 读取配置
	initial.DBInit()       // 初始化数据库
	initial.MemCacheInit() // 初始化缓存

	// 如果配置文件中开启了调试模式
	if !utility.IsDebugMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	router := initial.RouterInit()
	MountRoutes(router)

	// 启动服务器
	utility.StartServer(router, ":"+global.Config.GetString("server.port"))
}
