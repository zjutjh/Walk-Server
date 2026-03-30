package middleware

import (
	"github.com/gin-gonic/gin"
	jwtmiddleware "github.com/zjutjh/mygo/jwt/middleware"
)

func Auth() gin.HandlerFunc {
	return jwtmiddleware.Auth[string](true)
}
