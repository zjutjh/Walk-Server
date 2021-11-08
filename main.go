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

	router.GET("/api/v1/oauth", controller.Oauth) // 微信 Oauth 的起点接口
	api := router.Group("/api/v1", middleware.TimeValidity)
	{
		// Basic
		api.GET("/login", controller.Login) // 微信服务器的回调地址

		// Register
		register := api.Group("/register", middleware.RegisterJWTValidity)
		{
			register.POST("/student", controller.StudentRegister) // 在校生报名地址
			register.POST("/teacher", controller.TeacherRegister) // 教职工报名地址
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
			team.GET("/info", controller.GetTeamInfo)                        // 获取团队信息
			team.POST("/create", controller.CreateTeam)                      // 创建团队
			team.POST("/update", controller.UpdateTeam)                      // 修改队伍信息
			team.POST("/join", controller.JoinTeam)                          // 加入团队
			team.GET("/leave", controller.LeaveTeam)                         // 离开团队
			team.GET("/remove", controller.RemoveMember)                     // 移除队员
			team.GET("/disband", controller.DisbandTeam)                     // 解散团队
			team.GET("/submit", middleware.CanSubmit, controller.SubmitTeam) // 提交团队
			team.GET("/match", controller.RandomMatch)                       // 随机匹配
			team.GET("/rollback", controller.RollBackTeam)                   // 撤销提交
		}
	}

	// start server
	utility.StartServer(router, ":"+initial.Config.GetString("server.port"))
}
