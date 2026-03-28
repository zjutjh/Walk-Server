package register

import (
	adminapi "app/api/admin"
	userapi "app/api/user"
	"app/middleware"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/middleware/cors"
	"github.com/zjutjh/mygo/session"
	midsession "github.com/zjutjh/mygo/session/middleware"
	"github.com/zjutjh/mygo/swagger"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())
	router.Use(session.Pick())

	r := router.Group(routePrefix())
	{
		if slices.Contains([]string{config.AppEnvDev, config.AppEnvTest}, config.AppEnv()) {
			r.GET("/swagger.json", swagger.DocumentHandler(router))
		}
		r.POST("/admin/register", adminapi.RegisterAdminHandler())
		r.POST("/admin/auth", adminapi.AuthAdminHandler())
		r.POST("/admin/user/update", midsession.Auth[int](true), adminapi.UpdateUserHandler())
		r.POST("/admin/team/bind", midsession.Auth[int](true), adminapi.BindCodeHandler())
		r.POST("/admin/team/update", midsession.Auth[int](true), adminapi.UpdateTeamHandler())
		r.POST("/admin/team/violation/mark", midsession.Auth[int](true), adminapi.MarkTeamViolationHandler())
		r.POST("/admin/destination/confirm", midsession.Auth[int](true), adminapi.ConfirmDestinationHandler())
		r.POST("/admin/team/regroup", midsession.Auth[int](true), middleware.RequireSuperAdmin(), adminapi.RegroupHandler())
		r.GET("/admin/team/status", midsession.Auth[int](true), adminapi.GetTeamStatusHandler())
		r.GET("/admin/user/info/code", midsession.Auth[int](true), middleware.RequireSuperAdmin(), adminapi.GetUserInfoByScanHandler())
		r.GET("/admin/user/info", midsession.Auth[int](true), middleware.RequireSuperAdmin(), adminapi.GetUserInfoByIDHandler())

		adminGroup := r.Group("/admin")
		{
			adminGroup.POST("/register", api.RegisterAdminHandler())
			adminGroup.POST("/auth", api.AuthAdminHandler())

			authAdminGroup := adminGroup.Group("", midsession.Auth[int64](true))
			{
				authAdminGroup.POST("/destination/confirm", api.ConfirmDestinationHandler())

				userGroup := authAdminGroup.Group("/user")
				{
					userGroup.POST("/update", api.UpdateUserHandler())
					userGroup.GET("/info/code", middleware.NeedPerm("super"), api.GetUserInfoByScanHandler())
					userGroup.GET("/info", middleware.NeedPerm("super"), api.GetUserInfoByIDHandler())
				}

				teamGroup := authAdminGroup.Group("/team")
				{
					teamGroup.POST("/bind", api.BindCodeHandler())
					teamGroup.POST("/update", api.UpdateTeamHandler())
					teamGroup.POST("/regroup", middleware.NeedPerm("super"), api.RegroupHandler())
					teamGroup.GET("/status", api.GetTeamStatusHandler())
					teamGroup.POST("/violation/mark", api.MarkTeamViolationHandler())
				}
			}
		}

		// 注册业务逻辑接口
		dashboardGroup := r.Group("/dashboard", midsession.Auth[int64](true)) // go强类型断言，int不通过
		user := r.Group("/user")
		{
			user.GET("/wechat/login", userapi.WechatLoginHandler())

			auth := user.Group("")
			auth.Use(middleware.Auth())
			{
				auth.POST("/register/student", userapi.RegisterStudentHandler())
				auth.POST("/register/teacher", userapi.RegisterTeacherHandler())
				auth.POST("/register/alumnus", userapi.RegisterAlumnusHandler())

				auth.GET("/info", userapi.UserInfoHandler())
				auth.POST("/modify", userapi.UserModifyHandler())

				auth.POST("/team/create", userapi.TeamCreateHandler())
				auth.POST("/team/join", userapi.TeamJoinHandler())
				auth.GET("/team/info", userapi.TeamInfoHandler())
				auth.POST("/team/update", userapi.TeamUpdateHandler())
				auth.POST("/team/leave", userapi.TeamLeaveHandler())
				auth.DELETE("/team/disband", userapi.TeamDisbandHandler())
			}
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
}

func routeTest(r *gin.RouterGroup, router *gin.Engine) {
	// 测试接口，不要鉴权
	dashboardGroup := r.Group("/dashboard")
	{
		dashboardGroup.GET("/overview", dashboard.OverviewHandler())
		dashboardGroup.GET("/checkpoint", dashboard.CheckpointHandler())
		dashboardGroup.GET("/segment", dashboard.SegmentHandler())
		dashboardGroup.GET("/permission", dashboard.PermissionHandler())

		teamGroup := dashboardGroup.Group("/teams")
		{
			teamGroup.GET("", teams.TeamHandler())
			teamGroup.POST("/lost", teams.LostHandler())
			teamGroup.GET("/filter", teams.FilterHandler())
		}

		dashboardGroup.GET("/stats/route/all", stats.AllHandler())
		dashboardGroup.GET("/stats/route", stats.RouteHandler())
	}
}
