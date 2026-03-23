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
	"app/middleware"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())
	router.Use(session.Pick())

	r := router.Group(routePrefix())
	{
		routeBase(r, router)

		user := r.Group("/user")
		{
			// 注册用户端业务逻辑接口
			user.GET("/wechat/login", api.WechatLoginHandler())

			auth := user.Group("")
			auth.Use(middleware.Auth())
			{
				auth.POST("/register/student", api.RegisterStudentHandler())
				auth.POST("/register/teacher", api.RegisterTeacherHandler())
				auth.POST("/register/alumnus", api.RegisterAlumnusHandler())

				auth.GET("/info", api.UserInfoHandler())
				auth.POST("/modify", api.UserModifyHandler())

				auth.POST("/team/create", api.TeamCreateHandler())
				auth.POST("/team/join", api.TeamJoinHandler())
				auth.GET("/team/info", api.TeamInfoHandler())
				auth.POST("/team/update", api.TeamUpdateHandler())
				auth.GET("/team/leave", api.TeamLeaveHandler())
				auth.GET("/team/disband", api.TeamDisbandHandler())
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
