package register

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/middleware/cors"
	"github.com/zjutjh/mygo/swagger"

	"app/api"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())

	r := router.Group(routePrefix())
	{
		routeBase(r, router)

		// 注册业务逻辑接口

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

	r.POST("/admin/team/update", api.UpdateTeamHandler())
	r.POST("/admin/user/update", api.UpdateUserHandler())
	r.POST("/admin/team/bind",api.BindCodeHandler())
	r.POST("/admin/team/violation/mark",api.MarkTeamViolationHandler())
	r.POST("/admin/destination/confirm",api.ConfirmDestinationHandler())
	r.POST("/admin/team/regroup",api.RegroupHandler())
	r.GET("/admin/team/status",api.GetTeamStatusHandler())
	r.GET("/admin/user/info/code",api.GetUserInfoByScanHandler())
	r.GET("/admin/user/info",api.GetUserInfoByIDHandler())
}
