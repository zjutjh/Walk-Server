package repo

import (
	"context"
	"errors"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type PeopleRepo struct {
	db *gorm.DB
}

func NewPeopleRepo() *PeopleRepo {
	return &PeopleRepo{db: ndb.Pick()}
}

func NewPeopleRepoWithDB(db *gorm.DB) *PeopleRepo {
	return &PeopleRepo{db: db}
}

func (r *PeopleRepo) FindByOpenID(ctx context.Context, openID string) (*Person, error) {
	var person Person
	err := r.db.WithContext(ctx).Where("open_id = ?", openID).First(&person).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *PeopleRepo) FindByIdentity(ctx context.Context, identity string) (*Person, error) {
	var person Person
	err := r.db.WithContext(ctx).Where("identity = ?", identity).First(&person).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *PeopleRepo) FindByStuID(ctx context.Context, stuID string) (*Person, error) {
	if stuID == "" {
		return nil, nil
	}
	var person Person
	err := r.db.WithContext(ctx).Where("stu_id = ?", stuID).First(&person).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *PeopleRepo) Create(ctx context.Context, person *Person) error {
	return r.db.WithContext(ctx).Create(person).Error
}

func (r *PeopleRepo) UpdateByOpenID(ctx context.Context, openID string, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&Person{}).Where("open_id = ?", openID).Updates(updates).Error
}

func (r *PeopleRepo) UpdateByTeamID(ctx context.Context, teamID int64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&Person{}).Where("team_id = ?", teamID).Updates(updates).Error
}

func (r *PeopleRepo) ListByTeamID(ctx context.Context, teamID int64) ([]Person, error) {
	var people []Person
	err := r.db.WithContext(ctx).Where("team_id = ?", teamID).Order("role DESC, id ASC").Find(&people).Error
	return people, err
}
