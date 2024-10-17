package router

import (
	"walk-server/controller/admin"
	"walk-server/controller/basic"
	"walk-server/controller/message"
	"walk-server/controller/poster"
	"walk-server/controller/register"
	"walk-server/controller/team"
	"walk-server/controller/user"
	"walk-server/middleware"

	"github.com/gin-gonic/gin"
)

func MountRoutes(router *gin.Engine) {
	router.POST("/api/v1/redis2mysql", middleware.TokenRateLimiter, team.RedisToMysql) // 从 Redis 中导入数据到 MySQL
	api := router.Group("/api/v1", middleware.TokenRateLimiter)
	{
		if !gin.IsDebugging() {
			api.Use(middleware.TimeValidity)
		}

		// Basic
		api.GET("/oauth", basic.Oauth) // 微信 Oauth 的起点接口
		api.GET("/login", basic.Login) // 微信服务器的回调地址

		// Register
		registerApi := api.Group("/register", middleware.RegisterJWTValidity, middleware.PerRateLimiter)
		{
			registerApi.POST("/student", middleware.IsExpired, register.StudentRegister) // 在校生报名地址
			registerApi.POST("/teacher", middleware.IsExpired, register.TeacherRegister) // 教职工报名地址
			registerApi.POST("/alumnus", register.Login)                                 // 导入成员登录地址
		}

		// User
		userApi := api.Group("/user", middleware.IsRegistered, middleware.PerRateLimiter)
		{
			userApi.GET("/info", user.GetInfo)                             // 获取用户信息
			userApi.POST("/modify", middleware.IsExpired, user.ModifyInfo) // 修改用户信息
		}

		// Team
		teamApi := api.Group("/team", middleware.IsRegistered, middleware.PerRateLimiter)
		{
			if gin.IsDebugging() {
				teamApi.GET("/submit", team.SubmitTeam) // 提交团队
			} else {
				teamApi.GET("/submit", middleware.IsExpired, middleware.CanSubmit, team.SubmitTeam) // 提交团队
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
		messageApi := api.Group("/message", middleware.IsRegistered, middleware.PerRateLimiter)
		{
			messageApi.GET("/list", message.ListMessage)                            // 获取所有的消息
			messageApi.POST("/delete", middleware.IsExpired, message.DeleteMessage) // 读了消息以后删除消息
		}

		// 海报相关的 API
		picApi := api.Group("/poster", middleware.IsRegistered)
		{
			picApi.GET("/get", poster.GetPoster) // 获取海报
		}
	}

	adminApi := router.Group("/api/v1/admin", middleware.TokenRateLimiter)
	{
		adminApi.POST("/auth", admin.AuthByPassword)                                     // 微信登录
		adminApi.POST("/auth/auto", admin.WeChatLogin)                                   // 自动登录
		adminApi.POST("/auth/without", admin.AuthWithoutCode)                            // 测试登录
		adminApi.GET("/team/status", middleware.CheckAdmin, admin.GetTeam)               // 获取队伍信息
		adminApi.POST("/team/bind", middleware.CheckAdmin, admin.BindTeam)               // 绑定队伍
		adminApi.POST("/team/update", middleware.CheckAdmin, admin.UpdateTeamStatus)     // 更新队伍状态
		adminApi.POST("/team/user_status", middleware.CheckAdmin, admin.UserStatus)      // 更新用户状态
		adminApi.POST("/team/destination", middleware.CheckAdmin, admin.PostDestination) // 提交终点
		adminApi.POST("/team/secret", middleware.CheckAdmin, admin.BlockWithSecret)      // 通过密钥封禁接口
		adminApi.POST("/team/regroup", middleware.CheckAdmin, admin.Regroup)             // 重新分组
		adminApi.POST("/team/submit", middleware.CheckAdmin, admin.SubmitTeam)           // 提交团队
		adminApi.GET("/detail", admin.GetDetail)                                         // 获取路线人员详情
		adminApi.GET("/submit", admin.GetSubmitDetail)                                   // 获取报名人员列表

	}

}
