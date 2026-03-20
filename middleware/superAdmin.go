package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/session"

	"app/comm"
	"app/dao/repo"
)

func RequireSuperAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		adminID, err := session.GetIdentity[int64](ctx)
		if err != nil {
			adminIDInt, err := session.GetIdentity[int](ctx)
			if err != nil {
				reply.Fail(ctx, comm.CodeNotLoggedIn)
				ctx.Abort()
				return
			}
			adminID = int64(adminIDInt)
		}

		adminRepo := repo.NewAdminRepo()
		admin, err := adminRepo.FindByID(ctx, adminID)
		if err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Error("查询管理员权限失败")
			reply.Fail(ctx, comm.CodeUnknownError)
			ctx.Abort()
			return
		}
		if admin == nil || admin.Permission != comm.AdminPermissionSuper {
			reply.Fail(ctx, comm.CodePermissionDenied)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
