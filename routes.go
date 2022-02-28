package main

import (
	"walk-server/controller/basic"
	"walk-server/controller/register"
	"walk-server/controller/team"
	"walk-server/controller/user"
	"walk-server/middleware"

	"github.com/gin-gonic/gin"
)

func MountRoutes(router *gin.Engine) {
	api := router.Group("/api/v1", middleware.TimeValidity)
	{
		// Basic
		api.GET("/oauth", basic.Oauth) // 微信 Oauth 的起点接口
		api.GET("/login", basic.Login) // 微信服务器的回调地址

		// Register
		registerApi := api.Group("/register", middleware.RegisterJWTValidity)
		{
			registerApi.POST("/student", middleware.IsExpired, register.StudentRegister) // 在校生报名地址
			registerApi.POST("/teacher", middleware.IsExpired, register.TeacherRegister) // 教职工报名地址
		}

		// User
		userApi := api.Group("/user", middleware.IsRegistered)
		{
			userApi.GET("/info", user.GetInfo)                             // 获取用户信息
			userApi.POST("/modify", middleware.IsExpired, user.ModifyInfo) // 修改用户信息
		}

		// Team
		teamApi := api.Group("/team", middleware.IsRegistered)
		{
			teamApi.GET("/info", team.GetTeamInfo)                                              // 获取团队信息
			teamApi.POST("/random-list", team.GetRandomList)                                    // 随机获取开放随机组队的团队列表
			teamApi.POST("/random-join", middleware.IsExpired, team.RandomJoin)                 // 通过随机列表加入团队
			teamApi.POST("/transfer-captain", middleware.IsExpired, team.TransferLeader)         // 转让队长
			teamApi.POST("/create", middleware.IsExpired, team.CreateTeam)                      // 创建团队
			teamApi.POST("/update", middleware.IsExpired, team.UpdateTeam)                      // 修改队伍信息
			teamApi.POST("/join", middleware.IsExpired, team.JoinTeam)                          // 加入团队
			teamApi.GET("/leave", middleware.IsExpired, team.LeaveTeam)                         // 离开团队
			teamApi.GET("/remove", middleware.IsExpired, team.RemoveMember)                     // 移除队员
			teamApi.GET("/disband", middleware.IsExpired, team.DisbandTeam)                     // 解散团队
			teamApi.GET("/submit", middleware.IsExpired, middleware.CanSubmit, team.SubmitTeam) // 提交团队
			teamApi.GET("/rollback", middleware.IsExpired, team.RollBackTeam)                   // 撤销提交
		}

		// 事件相关的 API
		messageApi := api.Group("/message", middleware.IsExpired)
		{
			messageApi.GET("/list") // 获取所有的事件
		}
	}
}
