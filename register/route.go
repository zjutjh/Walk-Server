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
