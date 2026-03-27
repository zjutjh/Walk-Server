package register

import (
	api "app/api/user"
	"app/middleware"

	"github.com/gin-gonic/gin"
)

func routeUser(r *gin.RouterGroup) {
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
			auth.POST("/team/leave", api.TeamLeaveHandler())
			auth.DELETE("/team/disband", api.TeamDisbandHandler())
		}
	}
}
