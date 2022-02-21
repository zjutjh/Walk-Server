/*
 * Copyright (c) 2021 IInfo.
 */

package utility

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

//StartServer start server with specific port
func StartServer(router *gin.Engine, port string) {
	var srv *http.Server

	log.Println("Server starting at" + port)

	srv = &http.Server{
		Addr:    port,
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 5)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

// IsDebugMode 获取服务器是否在调试模式
func IsDebugMode() bool {
	initial.ConfigInit()
	return initial.Config.GetBool("server.debug")
}
