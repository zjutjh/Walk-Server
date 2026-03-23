package repo

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"app/dao/model"
	"app/dao/query"
)

type AdminRepo struct {
	query *query.Query
}

func NewAdminRepo() *AdminRepo {
	return &AdminRepo{
		query: newQuery(),
	}
}

// FindByID 根据ID查询管理员
func (r *AdminRepo) FindByID(ctx context.Context, id int64) (*model.Admin, error) {
	a := r.query.Admin
	record, err := a.WithContext(ctx).Where(a.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
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

// Create 创建管理员
func (r *AdminRepo) Create(ctx context.Context, admin *model.Admin) error {
	return r.query.Admin.WithContext(ctx).Create(admin)
}
