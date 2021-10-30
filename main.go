package main

import (
	"walk-server/controller"
	"walk-server/middleware"
	"walk-server/utility"
	"walk-server/utility/initial"
)

func main() {
	initial.ConfigInit() // 读取配置
	initial.DBInit()     // 初始化数据库
	router := initial.RouterInit()

	api := router.Group("/api/v1", middleware.TimeValidity)
	{
		// Basic
		api.GET("/login", controller.Login) // 微信服务器的回调地址
		api.GET("/oauth", controller.Oauth) // 微信 Oauth 的起点接口

		// Register
		register := api.Group("/register", middleware.RegisterJWTValidity)
		{
			register.POST("/student", controller.StudentRegister)   // 在校生报名地址
			register.POST("/graduate", controller.GraduateRegister) // 校友报名地址
		}

		// User
		user := api.Group("/user", middleware.Auth)
		{
			user.GET("/info", controller.GetInfo)       // 获取用户信息
			user.POST("/modify", controller.ModifyInfo) // 修改用户信息
		}

		// Team
		team := api.Group("/team", middleware.IsRegistered)
		{
			team.GET("/info")
			team.POST("/create", controller.CreateTeam) // 创建团队
			team.POST("/modify")
			team.POST("/join", controller.JoinTeam) // 加入团队
			team.GET("/leave")
			team.GET("/disband")
		}

	}

	// start server
	utility.StartServer(router, ":8080")
}
