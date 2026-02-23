package repo

import (
	"context"
	"errors"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
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

func (r *TeamRepo) UpdateByID(ctx context.Context, id int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&Team{}).Where("id = ?", id).Updates(updates).Error
}

func (r *TeamRepo) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Team{}).Error
}
