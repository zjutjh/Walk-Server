package register

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/middleware/cors"
	"github.com/zjutjh/mygo/swagger"

	"app/api"
	"app/middleware"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())

	r := router.Group(routePrefix())
	{
		routeBase(r, router)

		// 注册业务逻辑接口
		r.GET("/wechat/login", api.WechatLoginHandler())

		auth := r.Group("")
		auth.Use(middleware.Auth())
		{
			auth.POST("/register/student", api.RegisterStudentHandler())
			auth.POST("/register/teacher", api.RegisterTeacherHandler())
			auth.POST("/register/alumnus", api.RegisterAlumnusHandler())

			auth.GET("/user/info", api.UserInfoHandler())
			auth.POST("/user/modify", api.UserModifyHandler())

			auth.POST("/team/create", api.TeamCreateHandler())
			auth.POST("/team/join", api.TeamJoinHandler())
			auth.GET("/team/info", api.TeamInfoHandler())
			auth.POST("/team/update", api.TeamUpdateHandler())
			auth.GET("/team/leave", api.TeamLeaveHandler())
			auth.GET("/team/disband", api.TeamDisbandHandler())
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
