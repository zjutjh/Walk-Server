package router

import (
	"walk-server/controller/admin"
	"walk-server/controller/basic"
	"walk-server/controller/message"
	"walk-server/controller/register"
	"walk-server/controller/team"
	"walk-server/controller/user"
	"walk-server/middleware"

	"github.com/gin-gonic/gin"
)

func MountRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		if !gin.IsDebugging() {
			api.Use(middleware.TimeValidity)
		}

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
			if gin.IsDebugging() {
				teamApi.GET("/submit", middleware.TokenRateLimiter, team.SubmitTeam) // 提交团队
			} else {
				teamApi.GET("/submit", middleware.TokenRateLimiter, middleware.PerRateLimiter, middleware.IsExpired, middleware.CanSubmit, team.SubmitTeam) // 提交团队
			}

			teamApi.GET("/info", team.GetTeamInfo)                              // 获取团队信息
			teamApi.POST("/random-list", team.GetRandomList)                    // 随机获取开放随机组队的团队列表
			teamApi.POST("/random-join", middleware.IsExpired, team.RandomJoin) // 通过随机列表加入团队
			teamApi.POST("/create", middleware.IsExpired, team.CreateTeam)      // 创建团队
			teamApi.POST("/update", middleware.IsExpired, team.UpdateTeam)      // 修改队伍信息
			teamApi.POST("/join", middleware.IsExpired, team.JoinTeam)          // 加入团队
			teamApi.GET("/leave", middleware.IsExpired, team.LeaveTeam)         // 离开团队
			teamApi.GET("/remove", middleware.IsExpired, team.RemoveMember)     // 移除队员
			teamApi.GET("/disband", middleware.IsExpired, team.DisbandTeam)     // 解散团队
			teamApi.GET("/rollback", middleware.IsExpired, team.RollBackTeam)   // 撤销提交
		}

		// 事件相关的 API
		messageApi := api.Group("/message", middleware.IsRegistered)
		{
			messageApi.GET("/list", message.ListMessage)                            // 获取所有的消息
			messageApi.POST("/delete", middleware.IsExpired, message.DeleteMessage) // 读了消息以后删除消息
		}

		// 海报相关的 API
		// picApi := api.Group("/poster", middleware.IsRegistered)
		// {
		// 	picApi.GET("/get", poster.GetPoster) // 获取海报
		// }

		admin2 := api.Group("/admin")
		{
			admin2.POST("/auth", admin.AuthByPassword)
			admin2.POST("/auth/auto", admin.WeChatLogin)
			admin2.POST("/auth/without", admin.AuthWithoutCode)
			admin2.POST("/user/sd", middleware.CheckAdmin, admin.UserSD)
			admin2.POST("/user/sm", middleware.CheckAdmin, admin.UserSM)
			admin2.POST("/team/sm", middleware.CheckAdmin, admin.TeamSM)
			admin2.POST("/team/out", middleware.CheckAdmin, admin.UpdateTeam)
			admin2.GET("/team/status", middleware.CheckAdmin, admin.GetTeam)
		}
	}
}
