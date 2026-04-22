package middleware

import (
	"app/comm"
	repo "app/dao/repo"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zjutjh/mygo/foundation/reply"
)

// middleware 只做鉴权
var permissionRank = map[string]int{
	"external": 4,
	"internal": 3,
	"manager":  2,
	"super":    1,
}

func getPermRank(perm string) (int, bool) {
	rank, ok := permissionRank[perm]
	return rank, ok
}

// NeedPerm 判断当前管理员权限是否达到传入的最低权限。
// 例如最低权限为 internal 时，internal/manager/super 都会放行。
func NeedPerm(minPerm string) gin.HandlerFunc {
	minPerm = strings.ToLower(strings.TrimSpace(minPerm))
	return func(ctx *gin.Context) {
		admin, ok := repo.GetAdminInfo(ctx)
		if !ok {
			return
		}

		currentPerm := strings.ToLower(strings.TrimSpace(admin.Permission))
		currentRank, currentOK := getPermRank(currentPerm)
		minRank, minOK := getPermRank(minPerm)

		if !currentOK || !minOK {
			reply.Fail(ctx, comm.CodeMiddlewareServiceError)
			return
		}

		if currentRank > minRank {
			reply.Fail(ctx, comm.CodePermissionDenied)
			return
		}
		ctx.Next()
	}
}
