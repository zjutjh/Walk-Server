package main

import (
	"walk-server/controller"
	"walk-server/middleware"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func main() {
	initial.ConfigInit() // 读取配置
	initial.DBInit()     // 初始化数据库

	// 如果配置文件中开启了调试模式
	if !utility.IsDebugMode() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := initial.RouterInit()
	api := router.Group("/api/v1", middleware.TimeValidity)
	{
		// Basic
		api.GET("/oauth", controller.Oauth) // 微信 Oauth 的起点接口
		api.GET("/login", controller.Login) // 微信服务器的回调地址

		// Register
		register := api.Group("/register", middleware.RegisterJWTValidity)
		{
			register.POST("/student", middleware.IsExpired, controller.StudentRegister) // 在校生报名地址
			register.POST("/teacher", middleware.IsExpired, controller.TeacherRegister) // 教职工报名地址
		}

		// User
		user := api.Group("/user", middleware.Auth)
		{
			user.GET("/info", controller.GetInfo)                             // 获取用户信息
			user.POST("/modify", middleware.IsExpired, controller.ModifyInfo) // 修改用户信息
		}

		// Team
		team := api.Group("/team", middleware.IsRegistered)
		{
			team.GET("/info", controller.GetTeamInfo)                                              // 获取团队信息
			team.POST("/create", middleware.IsExpired, controller.CreateTeam)                      // 创建团队
			team.POST("/update", middleware.IsExpired, controller.UpdateTeam)                      // 修改队伍信息
			team.POST("/join", middleware.IsExpired, controller.JoinTeam)                          // 加入团队
			team.GET("/leave", middleware.IsExpired, controller.LeaveTeam)                         // 离开团队
			team.GET("/remove", middleware.IsExpired, controller.RemoveMember)                     // 移除队员
			team.GET("/disband", middleware.IsExpired, controller.DisbandTeam)                     // 解散团队
			team.GET("/submit", middleware.IsExpired, middleware.CanSubmit, controller.SubmitTeam) // 提交团队
			team.GET("/match", middleware.IsExpired, controller.RandomMatch)                       // 随机匹配
			team.GET("/rollback", middleware.IsExpired, controller.RollBackTeam)                   // 撤销提交
		}
	}

	// start server
	utility.StartServer(router, ":"+initial.Config.GetString("server.port"))
}
