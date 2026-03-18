package middleware

import (
	"app/comm"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"

	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/session"
)

//下面两个还是放在service里面，middleawar只做鉴权
//但是考虑到没有service，而dao的cache意义不明，还是留在这里吧（

// 从session反查id
func GetAdminID(ctx *gin.Context) (int64, bool) {
	adminID, err := session.GetIdentity[int64](ctx)
	if err != nil {
		reply.Fail(ctx, comm.CodeNotLoggedIn)
		return 0, false
	}
	return adminID, true
}

// GetAdmin 获取当前登录的管理员信息
func GetAdminInfo(ctx *gin.Context) (*model.Admin, bool) {
	adminID, ok := GetAdminID(ctx)
	if !ok {
		return nil, false
	}

	adminRepo := repo.NewAdminRepo()
	//这里应该可以优化，我只需要权限 信息即可
	admin, err := adminRepo.FindByID(ctx, adminID)
	if err != nil {
		reply.Fail(ctx, comm.CodeUnknownError)
		return nil, false
	}
	if admin == nil {
		reply.Fail(ctx, comm.CodeNotLoggedIn)
		return nil, false
	}

	return admin, true
}

// 以下是查询权限
var permissionRank = map[string]int{
	"external": 4,
	"internal": 3,
	"manager":  2,
	"super":    1,
}

// func normalizePerm(perm string) string {
// 	return strings.ToLower(strings.TrimSpace(perm))
// }

func getPermRank(perm string) (int, bool) {
	rank, err := permissionRank[perm]
	return rank, err
}

// NeedPerm 判断当前管理员权限是否达到传入的最低权限。
// 例如最低权限为 internal 时，internal/manager/super 都会放行。
func NeedPerm(minPerm string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		admin, ok := GetAdminInfo(ctx)
		if !ok {
			return
		}

		currentRank, currentOK := getPermRank(admin.Permission)
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
