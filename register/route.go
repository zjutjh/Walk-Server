package register

import (
	api "app/api/admin"
	"app/api/dashboard"
	"app/api/dashboard/stats"
	"app/api/dashboard/teams"
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
		// routeTest(r, router)

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
		{
			dashboardGroup.GET("/overview", middleware.NeedPerm("internal"), dashboard.OverviewHandler())
			dashboardGroup.GET("/checkpoint", middleware.NeedPerm("internal"), dashboard.CheckpointHandler())
			dashboardGroup.GET("/segment", middleware.NeedPerm("internal"), dashboard.SegmentHandler())
			dashboardGroup.GET("/permission", dashboard.PermissionHandler()) // 不用限制权限等级

			teamGroup := dashboardGroup.Group("/teams")
			{
				teamGroup.GET("", middleware.NeedPerm("manager"), teams.TeamHandler())
				teamGroup.POST("/lost", middleware.NeedPerm("manager"), teams.LostHandler())
				teamGroup.GET("/filter", middleware.NeedPerm("internal"), teams.FilterHandler())
			}

			dashboardGroup.GET("/stats/route/all", middleware.NeedPerm("internal"), stats.AllHandler())
			dashboardGroup.GET("/stats/route", middleware.NeedPerm("internal"), stats.RouteHandler())
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
