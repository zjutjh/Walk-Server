package register

import (
	api "app/api/admin"
	"app/middleware"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	midsession "github.com/zjutjh/mygo/session/middleware"
	"github.com/zjutjh/mygo/swagger"
)

func routeAdmin(r *gin.RouterGroup, router *gin.Engine) {
	// OpenAPI/Swagger 文档生成
	if slices.Contains([]string{config.AppEnvDev, config.AppEnvTest}, config.AppEnv()) {
		r.GET("/swagger.json", swagger.DocumentHandler(router))
	}
	r.POST("/admin/register", api.RegisterAdminHandler())
	r.POST("/admin/auth", api.AuthAdminHandler())
	r.POST("/admin/user/update", midsession.Auth[int](true), api.UpdateUserHandler())
	r.POST("/admin/team/bind", midsession.Auth[int](true), api.BindCodeHandler())
	r.POST("/admin/team/update", midsession.Auth[int](true), api.UpdateTeamHandler())
	r.POST("/admin/team/violation/mark", midsession.Auth[int](true), api.MarkTeamViolationHandler())
	r.POST("/admin/destination/confirm", midsession.Auth[int](true), api.ConfirmDestinationHandler())
	r.POST("/admin/team/regroup", midsession.Auth[int](true), middleware.RequireSuperAdmin(), api.RegroupHandler())
	r.GET("/admin/team/status", midsession.Auth[int](true), api.GetTeamStatusHandler())
	r.GET("/admin/user/info/code", midsession.Auth[int](true), middleware.RequireSuperAdmin(), api.GetUserInfoByScanHandler())
	r.GET("/admin/user/info", midsession.Auth[int](true), middleware.RequireSuperAdmin(), api.GetUserInfoByIDHandler())
}
