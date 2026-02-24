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

		dashboardGroup := r.Group("/dashboard")
		{
			dashboardGroup.GET("/overview", dashboard.OverviewHandler())
			dashboardGroup.GET("/checkpoint", dashboard.CheckpointHandler())
			dashboardGroup.GET("/segment", dashboard.SegmentHandler())
			dashboardGroup.GET("/permission", dashboard.PermissionHandler())

			teamGroup := dashboardGroup.Group("/teams")
			{
				teamGroup.GET(":team_id", dashboard.TeamHandler())
				teamGroup.GET("/filter", teams.FilterHandler())
			}

			statsGroup := dashboardGroup.Group("/stats/route")
			{
				statsGroup.GET("/all", stats.AllHandler())
				statsGroup.GET("/", stats.RouteHandler())
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

	// 健康检查
	r.GET("/health", api.HealthHandler())
}
