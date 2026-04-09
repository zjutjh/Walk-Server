package repo

import (
	"app/comm"
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/session"
	"gorm.io/gorm"

	adminCache "app/dao/cache/admin"
	"app/dao/model"
	"app/dao/query"
)

type AdminRepo struct {
	query *query.Query
}

func NewAdminRepo() *AdminRepo {
	return &AdminRepo{
		query: query.Use(ndb.Pick()),
	}
}

// FindByID 根据ID查询管理员
func (r *AdminRepo) FindByID(ctx context.Context, id int64) (*model.Admin, error) {
	if record, hit, err := adminCache.GetAdmin(ctx, id); err == nil && hit {
		return record, nil
	}

	a := r.query.Admin
	record, err := a.WithContext(ctx).Where(a.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = adminCache.SetAdmin(ctx, record)
	return record, nil
}

// FindByAccount 根据账号查询管理员
func (r *AdminRepo) FindByAccount(ctx context.Context, account string) (*model.Admin, error) {
	a := r.query.Admin
	record, err := a.WithContext(ctx).Where(a.Account.Eq(account)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// 从session反查id
func GetAdminID(ctx *gin.Context) (int64, bool) {
	// if !hasSessionCookie(ctx) {
	// 	reply.Fail(ctx, comm.CodeNotLoggedIn)
	// 	return 0, false
	// }
	adminID, err := session.GetIdentity[int64](ctx)
	//fmt.Println("middleware get adminID:", adminID)
	//fmt.Println("err:", err)

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

	adminRepo := NewAdminRepo()

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
