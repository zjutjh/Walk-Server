package register

import (
	adminapi "app/api/admin"
	"app/api/dashboard"
	"app/api/dashboard/stats"
	"app/api/dashboard/teams"
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
		routeBase(r, router)

		adminGroup := r.Group("/admin")
		{
			adminGroup.POST("/register", adminapi.RegisterAdminHandler())
			adminGroup.POST("/auth", adminapi.AuthAdminHandler())

			authAdminGroup := adminGroup.Group("", midsession.Auth[int64](true))
			{
				authAdminGroup.POST("/destination/confirm", adminapi.ConfirmDestinationHandler())

				userGroup := authAdminGroup.Group("/user")
				{
					userGroup.POST("/update", adminapi.UpdateUserHandler())
					userGroup.GET("/info/code", middleware.NeedPerm("super"), adminapi.GetUserInfoByScanHandler())
					userGroup.GET("/info", middleware.NeedPerm("super"), adminapi.GetUserInfoByIDHandler())
				}

				teamGroup := authAdminGroup.Group("/team")
				{
					teamGroup.POST("/bind", adminapi.BindCodeHandler())
					teamGroup.POST("/update", adminapi.UpdateTeamHandler())
					teamGroup.POST("/regroup", middleware.NeedPerm("super"), adminapi.RegroupHandler())
					teamGroup.GET("/status", adminapi.GetTeamStatusHandler())
					teamGroup.POST("/violation/mark", adminapi.MarkTeamViolationHandler())
				}
			}
		}

<<<<<<< HEAD
		dashboardGroup := r.Group("/dashboard", midsession.Auth[int64](true))
=======
		// 注册业务逻辑接口
		dashboardGroup := r.Group("/dashboard", midsession.Auth[int64](true)) // go强类型断言，int不通过
		user := r.Group("/user")
>>>>>>> origin/dev
		{
			dashboardGroup.GET("/overview", middleware.NeedPerm("internal"), dashboard.OverviewHandler())
			dashboardGroup.GET("/checkpoint", middleware.NeedPerm("internal"), dashboard.CheckpointHandler())
			dashboardGroup.GET("/segment", middleware.NeedPerm("internal"), dashboard.SegmentHandler())
			dashboardGroup.GET("/permission", dashboard.PermissionHandler())

			teamGroup := dashboardGroup.Group("/teams")
			{
				teamGroup.GET("", middleware.NeedPerm("manager"), teams.TeamHandler())
				teamGroup.POST("/lost", middleware.NeedPerm("manager"), teams.LostHandler())
				teamGroup.GET("/filter", middleware.NeedPerm("internal"), teams.FilterHandler())
			}

			dashboardGroup.GET("/stats/route/all", middleware.NeedPerm("internal"), stats.AllHandler())
			dashboardGroup.GET("/stats/route", middleware.NeedPerm("internal"), stats.RouteHandler())
		}

		userGroup := r.Group("/user")
		{
			userGroup.GET("/wechat/login", userapi.WechatLoginHandler())

			authUserGroup := userGroup.Group("")
			authUserGroup.Use(middleware.Auth())
			{
				authUserGroup.POST("/register/student", userapi.RegisterStudentHandler())
				authUserGroup.POST("/register/teacher", userapi.RegisterTeacherHandler())
				authUserGroup.POST("/register/alumnus", userapi.RegisterAlumnusHandler())

				authUserGroup.GET("/info", userapi.UserInfoHandler())
				authUserGroup.POST("/modify", userapi.UserModifyHandler())

				authUserGroup.POST("/team/create", userapi.TeamCreateHandler())
				authUserGroup.POST("/team/join", userapi.TeamJoinHandler())
				authUserGroup.GET("/team/info", userapi.TeamInfoHandler())
				authUserGroup.POST("/team/update", userapi.TeamUpdateHandler())
				authUserGroup.POST("/team/leave", userapi.TeamLeaveHandler())
				authUserGroup.DELETE("/team/disband", userapi.TeamDisbandHandler())
			}
		}
	}
}

func routePrefix() string {
	return "/api"
}

func routeBase(r *gin.RouterGroup, router *gin.Engine) {
	if slices.Contains([]string{config.AppEnvDev, config.AppEnvTest}, config.AppEnv()) {
		r.GET("/swagger.json", swagger.DocumentHandler(router))
	}
}
<<<<<<< HEAD
=======

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
>>>>>>> origin/dev
