package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"

	"app/comm"
)

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorization := ctx.GetHeader("Authorization")
		if authorization == "" {
			reply.Fail(ctx, comm.CodeNotLoggedIn)
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authorization, "Bearer ")
		if tokenString == authorization {
			reply.Fail(ctx, comm.CodeNotLoggedIn)
			ctx.Abort()
			return
		}

		claims, err := comm.ParseToken(tokenString)
		if err != nil || claims == nil || claims.OpenID == "" {
			reply.Fail(ctx, comm.CodeLoginExpired)
			ctx.Abort()
			return
		}

		ctx.Set("open_id", claims.OpenID)
		ctx.Next()
	}
}
