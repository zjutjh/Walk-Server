package repo

import (
	"context"
	"errors"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TeamRepo struct {
	db *gorm.DB
}

func NewTeamRepo() *TeamRepo {
	return &TeamRepo{db: ndb.Pick()}
}

func NewTeamRepoWithDB(db *gorm.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) Create(ctx context.Context, team *Team) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *TeamRepo) FindByID(ctx context.Context, id int64) (*Team, error) {
	var team Team
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepo) FindByName(ctx context.Context, name string) (*Team, error) {
	var team Team
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepo) FindByNameExceptID(ctx context.Context, name string, id int64) (*Team, error) {
	var team Team
	err := r.db.WithContext(ctx).Where("name = ? AND id <> ?", name, id).First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepo) UpdateByID(ctx context.Context, id int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&Team{}).Where("id = ?", id).Updates(updates).Error
}

func (r *TeamRepo) IncrementNumIfAvailable(ctx context.Context, id int64, maxTeamSize int) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&Team{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND submit = ? AND num < ?", id, false, maxTeamSize).
		UpdateColumn("num", gorm.Expr("num + ?", 1))
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DecrementNumIfPositive(ctx context.Context, id int64) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&Team{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND submit = ? AND num > 0", id, false).
		UpdateColumn("num", gorm.Expr("num - ?", 1))
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Team{}).Error
}
