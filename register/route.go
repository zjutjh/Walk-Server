package register

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/jwt/middleware"
	"github.com/zjutjh/mygo/middleware/cors"
	"github.com/zjutjh/mygo/swagger"

	"app/api"
	"app/api/register"
	"app/api/team"
	"app/api/user"
	"app/api/wechat"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())

	r := router.Group(routePrefix())
	{
		routeBase(r, router)

		// 注册业务逻辑接口
		wx := r.Group("/wechat")
		{
			wx.GET("/oauth", wechat.OauthHandler())
			wx.GET("/login", wechat.LoginHandler())
			wx.GET("/miniprogram/login", wechat.MiniProgramLoginHandler())

			wxAuth := wx.Group("")
			wxAuth.Use(middleware.Auth(true))
			{
				wxAuth.POST("/message", wechat.SendMessageHandler())
			}
		}

		// Register - 报名
		reg := r.Group("/register")
		reg.Use(middleware.Auth(true))
		{
			reg.POST("/student", register.StudentRegisterHandler())
			reg.POST("/teacher", register.TeacherRegisterHandler())
		}

		// User - 用户相关
		usr := r.Group("/user")
		usr.Use(middleware.Auth(true))
		{
			usr.GET("/info", user.GetInfoHandler())
			usr.POST("/modify", user.ModifyInfoHandler())
		}

		// Team - 队伍相关
		tm := r.Group("/team")
		tm.Use(middleware.Auth(true))
		{
			tm.POST("/create", team.CreateTeamHandler())
			tm.POST("/join", team.JoinTeamHandler())
			tm.GET("/info", team.GetTeamInfoHandler())
			tm.POST("/leave", team.LeaveTeamHandler())
			tm.POST("/disband", team.DisbandTeamHandler())
			tm.POST("/remove", team.RemoveMemberHandler())
			tm.POST("/update", team.UpdateTeamHandler())
			tm.GET("/random-list", team.GetRandomListHandler())
			tm.POST("/random-join", team.JoinRandomHandler())
			tm.POST("/captain", team.TransferCaptainHandler())
			tm.POST("/add", team.AddMemberHandler())
		}
	}
}

func routePrefix() string {
	return "/api"
}

func routeBase(r *gin.RouterGroup, router *gin.Engine) {
	// OpenAPI/Swagger 文档生成
	if slices.Contains([]string{config.AppEnvDev, config.AppEnvTest}, config.AppEnv()) {
		r.GET("/swagger.json", swagger.DocumentHandler(router))
	}

	// 健康检查
	r.GET("/health", api.HealthHandler())
}
