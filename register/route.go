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
	r.POST("/admin/register", api.RegisterAdminHandler())
	r.POST("/admin/auth", api.AuthAdminHandler())
	r.POST("/admin/user/update", midsession.Auth[int64](true), api.UpdateUserHandler())
	r.POST("/admin/team/bind", midsession.Auth[int64](true), api.BindCodeHandler())
	r.POST("/admin/team/update", midsession.Auth[int64](true), api.UpdateTeamHandler())
	r.POST("/admin/team/violation/mark", midsession.Auth[int64](true), api.MarkTeamViolationHandler())
	r.POST("/admin/destination/confirm", midsession.Auth[int64](true), api.ConfirmDestinationHandler())
	r.POST("/admin/team/regroup", midsession.Auth[int64](true), middleware.RequireSuperAdmin(), api.RegroupHandler())
	r.GET("/admin/team/status", midsession.Auth[int64](true), api.GetTeamStatusHandler())
	r.GET("/admin/user/info/code", midsession.Auth[int64](true), middleware.RequireSuperAdmin(), api.GetUserInfoByScanHandler())
	r.GET("/admin/user/info", midsession.Auth[int64](true), middleware.RequireSuperAdmin(), api.GetUserInfoByIDHandler())
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
