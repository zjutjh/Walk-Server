package repo

import (
	"context"
	"errors"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"

	"app/dao/model"
)

type TeamRepo struct {
	db *gorm.DB
}

func NewTeamRepo() *TeamRepo {
	return &TeamRepo{
		db: ndb.Pick(),
	}
}

// FindById 根据ID查询队伍
func (r *TeamRepo) FindById(ctx context.Context, id int64) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).First(&team, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// FindByName 根据名称查询队伍
func (r *TeamRepo) FindByName(ctx context.Context, name string) (*model.Team, error) {
	var team model.Team
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// CheckNameExistsExcludingId 检查名称是否存在（排除指定ID）
func (r *TeamRepo) CheckNameExistsExcludingId(ctx context.Context, name string, excludeId int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Team{}).Where("name = ? AND id != ?", name, excludeId).Count(&count).Error
	return count > 0, err
}

// Create 创建队伍
func (r *TeamRepo) Create(ctx context.Context, tx *gorm.DB, team *model.Team) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(team).Error
}

// Delete 删除队伍
func (r *TeamRepo) Delete(ctx context.Context, tx *gorm.DB, id int64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Delete(&model.Team{}, id).Error
}

// Save 保存/更新队伍
func (r *TeamRepo) Save(ctx context.Context, tx *gorm.DB, team *model.Team) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Save(team).Error
}

// Update 更新队伍信息
func (r *TeamRepo) Update(ctx context.Context, tx *gorm.DB, team *model.Team) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Save(team).Error
}

// GetRandomList 获取随机匹配列表
func (r *TeamRepo) GetRandomList(ctx context.Context, limit int) ([]model.Team, error) {
	var teams []model.Team
	// Find teams that allow match and are not full (assuming max 6)
	err := r.db.WithContext(ctx).Where("allow_match = ? AND num < ?", true, 6).Limit(limit).Find(&teams).Error
	return teams, err
}

// FindByIdForUpdate 锁行查询
func (r *TeamRepo) FindByIdForUpdate(ctx context.Context, tx *gorm.DB, id int64) (*model.Team, error) {
	var team model.Team
	if tx == nil {
		return nil, errors.New("transaction required")
	}
	if err := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").First(&team, id).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

// Transaction 开启事务
func (r *TeamRepo) Transaction(ctx context.Context, fc func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fc)
}
