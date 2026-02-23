package register

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/middleware/cors"
	"github.com/zjutjh/mygo/swagger"

	"app/api"
	"app/api/dashboard"
	"app/api/dashboard/stats"
	"app/api/dashboard/teams"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())

	r := router.Group(routePrefix())
	{
		routeBase(r, router)

		// 注册业务逻辑接口

		r.GET("/dashboard/overview", dashboard.OverviewHandler())
		r.GET("/dashboard/checkpoint", dashboard.CheckpointHandler())
		r.GET("/dashboard/segment", dashboard.SegmentHandler())
		r.GET("/dashboard/teams/filter", teams.FilterHandler())
		r.GET("/dashboard/team/:team_id", teams.TeamHandler())
		r.GET("/dashboard/stats/route/all", stats.AllHandler())
		r.GET("/dashboard/stats/route", stats.RouteHandler())
		r.GET("/dashboard/permission", dashboard.PermissionHandler())
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
