package register

import (
	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/middleware/cors"
	"github.com/zjutjh/mygo/session"
)

func Route(router *gin.Engine) {
	router.Use(cors.Pick())
	router.Use(session.Pick())

	r := router.Group(routePrefix())
	{
		routeAdmin(r, router)
		routeUser(r)
	}
}

func routePrefix() string {
	return "/api"
}
