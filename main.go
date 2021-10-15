package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"walk-server/handler"
)

func main() {
	router := handler.InitRouter()

	router.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// start server
	handler.StartServer(router, ":8080")
}
