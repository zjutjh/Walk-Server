package register

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/middleware/cors"
	"github.com/zjutjh/mygo/session"
	midsession "github.com/zjutjh/mygo/session/middleware"
	"github.com/zjutjh/mygo/swagger"

	"app/api"
	"app/api/dashboard"
	"app/api/dashboard/stats"
	"app/api/dashboard/teams"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())
	router.Use(session.Pick())

	r := router.Group(routePrefix())
	{
		routeBase(r, router)

		// 注册业务逻辑接口

		dashboardGroup := r.Group("/dashboard")
		{
			dashboardGroup.GET("/overview", dashboard.OverviewHandler())
			dashboardGroup.GET("/checkpoint", dashboard.CheckpointHandler())
			dashboardGroup.GET("/segment", dashboard.SegmentHandler())
			dashboardGroup.GET("/permission", dashboard.PermissionHandler())

			teamGroup := dashboardGroup.Group("/teams")
			{
				teamGroup.GET("", teams.TeamHandler())
				teamGroup.GET("/filter", teams.FilterHandler())
				teamGroup.POST("/lost", teams.LostHandler())
			}

			dashboardGroup.GET("/stats/route/all", stats.AllHandler())
			dashboardGroup.GET("/stats/route", stats.RouteHandler())
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
	r.POST("/admin/register",api.RegisterAdminHandler())
	r.POST("/admin/auth", api.AuthAdminHandler())
	r.POST("/admin/user/update", midsession.Auth[int](true), api.UpdateUserHandler())
	r.POST("/admin/team/bind",midsession.Auth[int](true),api.BindCodeHandler())
	r.POST("/admin/team/violation/mark",midsession.Auth[int](true),api.MarkTeamViolationHandler())
	r.POST("/admin/destination/confirm",midsession.Auth[int](true),api.ConfirmDestinationHandler())
	r.POST("/admin/team/regroup",midsession.Auth[int](true),api.RegroupHandler())
	r.GET("/admin/team/status",midsession.Auth[int](true),api.GetTeamStatusHandler())
	r.GET("/admin/user/info/code",midsession.Auth[int](true),api.GetUserInfoByScanHandler())
	r.GET("/admin/user/info",midsession.Auth[int](true),api.GetUserInfoByIDHandler())
}
