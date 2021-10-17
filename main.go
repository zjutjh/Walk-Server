package main

import (
	"walk-server/controller"
	"walk-server/handler"
	"walk-server/middleware"
)

func main() {
	router := handler.InitRouter()

	api := router.Group("/api/v1", middleware.TimeValidity)
	{
		// Basic
		api.GET("/login", controller.Login)
		api.POST("/register", controller.Register)

		// User
		user := api.Group("/user", middleware.Auth)
		{
			user.GET("/info")
			user.POST("/modify")
		}

		// Team
		team := api.Group("/team", middleware.Auth)
		{

			team.GET("/info")
			team.GET("/create")
			team.POST("/modify")
			team.GET("/join")
			team.GET("/leave")
			team.GET("/disband")
		}

	}

	// start server
	handler.StartServer(router, ":8080")
}
