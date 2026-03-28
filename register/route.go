package register

import (
	api "app/api/admin"
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
