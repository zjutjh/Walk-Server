package register

import (
	api "app/api/admin"
	"app/api/dashboard"
	"app/api/dashboard/stats"
	"app/api/dashboard/teams"
	basicapi "app/api/user/basic"
	registerapi "app/api/user/register"
	teamapi "app/api/user/team"
	userapi "app/api/user/user"
	"app/middleware"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/config"
	jwtmiddleware "github.com/zjutjh/mygo/jwt/middleware"
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
		adminGroup := r.Group("/admin")
		{
			adminGroup.POST("/auth", api.AuthAdminHandler())

			authAdminGroup := adminGroup.Group("", midsession.Auth[int64](true))
			{
				authAdminGroup.POST("/logout", api.LogoutAdminHandler())
				authAdminGroup.POST("/destination/confirm", api.ConfirmDestinationHandler())

				userGroup := authAdminGroup.Group("/user")
				{
					userGroup.POST("/update", api.UpdateUserHandler())
					userGroup.POST("/violation/mark", api.MarkUserViolationHandler())
					userGroup.POST("/pending/start", middleware.NeedPerm("super"), api.StartPendingPeopleHandler())
					userGroup.GET("/info", middleware.NeedPerm("super"), api.GetUserInfoByIDHandler())
				}

				teamGroup := authAdminGroup.Group("/team")
				{
					teamGroup.POST("/bind", api.BindCodeHandler())
					teamGroup.POST("/update", api.UpdateTeamHandler())
					teamGroup.POST("/rebuild", middleware.NeedPerm("super"), api.RebuildHandler())
					teamGroup.GET("/status", api.GetTeamStatusHandler())
					teamGroup.POST("/violation/mark", api.MarkTeamViolationHandler())
				}
			}
		}

		// 注册业务逻辑接口
		dashboardGroup := r.Group("/dashboard", midsession.Auth[int64](true)) // go强类型断言，int不通过
		{
			dashboardGroup.GET("/overview", middleware.NeedPerm("internal"), dashboard.OverviewHandler())
			dashboardGroup.GET("/checkpoint", middleware.NeedPerm("internal"), dashboard.CheckpointHandler())
			dashboardGroup.GET("/segment", middleware.NeedPerm("internal"), dashboard.SegmentHandler())
			dashboardGroup.GET("/permission", dashboard.PermissionHandler())

			teamGroup := dashboardGroup.Group("/teams")
			{
				teamGroup.GET("", middleware.NeedPerm("manager"), teams.TeamHandler())
				teamGroup.POST("/lost", middleware.NeedPerm("manager"), teams.LostHandler())
				teamGroup.GET("/filter", middleware.NeedPerm("internal"), teams.FilterHandler())
			}

			dashboardGroup.GET("/stats/route/all", middleware.NeedPerm("internal"), stats.AllHandler())
			dashboardGroup.GET("/stats/route", middleware.NeedPerm("internal"), stats.RouteHandler())
		}
		user := r.Group("/user")
		{
			user.POST("/login", basicapi.LoginHandler())
			user.POST("/register/student", registerapi.RegisterStudentHandler())
			user.POST("/register/teacher", registerapi.RegisterTeacherHandler())
			user.POST("/register/alumnus", registerapi.RegisterAlumnusHandler())

			auth := user.Group("")
			auth.Use(jwtmiddleware.Auth[string](true))
			{
				auth.GET("/info", userapi.UserInfoHandler())
				auth.POST("/modify", userapi.UserModifyHandler())

				auth.POST("/team/create", teamapi.TeamCreateHandler())
				auth.POST("/team/join", teamapi.TeamJoinHandler())
				auth.GET("/team/overview", teamapi.TeamOverviewHandler())
				auth.GET("/team/detail", teamapi.TeamDetailHandler())
				auth.GET("/team/member", teamapi.TeamMemberHandler())
				auth.POST("/team/submit", teamapi.TeamSubmitHandler())
				auth.POST("/team/rollback", teamapi.TeamRollbackHandler())
				auth.GET("/team/random-list", teamapi.TeamRandomListHandler())
				auth.POST("/team/random-join", teamapi.TeamRandomJoinHandler())
				auth.POST("/team/update", teamapi.TeamUpdateHandler())
				auth.POST("/team/remove", teamapi.TeamRemoveMemberHandler())
				auth.POST("/team/captain", teamapi.TeamChangeCaptainHandler())
				auth.POST("/team/leave", teamapi.TeamLeaveHandler())
				auth.POST("/team/disband", teamapi.TeamDisbandHandler())

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
}
