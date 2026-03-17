package middleware

import (
	"app/comm"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"

	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/session"
)

func GetAdminID(ctx *gin.Context) (int64, bool) {
	adminID, err := session.GetIdentity[int64](ctx)
	if err != nil {
		reply.Fail(ctx, comm.CodeNotLoggedIn)
		return 0, false
	}
	return adminID, true
}

// GetAdmin 获取当前登录的管理员信息
// TODO: 写入gin.C
func GetAdmin(ctx *gin.Context) (*model.Admin, bool) {
	adminID, ok := GetAdminID(ctx)
	if !ok {
		return nil, false
	}

	adminRepo := repo.NewAdminRepo()
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
