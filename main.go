package main

import (
	"walk-server/controller"
	"walk-server/middleware"
	"walk-server/utility"
	"walk-server/utility/initial"
)

func main() {
	initial.ConfigInit() // 读取配置
	router := initial.RouterInit()

	api := router.Group("/api/v1", middleware.TimeValidity)
	{
		// Basic
		api.GET("/login", controller.Login)        // 微信服务器的回调地址
		api.POST("/register", controller.Register) // 报名地址
		api.GET("/oauth", controller.Oauth)        // 微信 Oauth 的起点接口

		// User
		user := api.Group("/user", middleware.Auth)
		{
			user.GET("/info")
			user.POST("/modify")
		}

		// Team
		team := api.Group("/team", middleware.Auth)
		{

			team.GET("/info")
			team.GET("/create")
			team.POST("/modify")
			team.GET("/join")
			team.GET("/leave")
			team.GET("/disband")
		}

	}

	// start server
	utility.StartServer(router, ":8080")
}
