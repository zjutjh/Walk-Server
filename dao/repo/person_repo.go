package repo

import (
	"context"
	"errors"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"

	"app/dao/model"
)

type PersonRepo struct {
	db *gorm.DB
}

func NewPersonRepo() *PersonRepo {
	return &PersonRepo{
		db: ndb.Pick(),
	}
}

// FindById 根据ID查询
func (r *PersonRepo) FindById(ctx context.Context, id int64) (*model.Person, error) {
	var person model.Person
	err := r.db.WithContext(ctx).First(&person, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &person, nil
}

// FindByStuId 根据学号查询
func (r *PersonRepo) FindByStuId(ctx context.Context, stuId string) (*model.Person, error) {
	var person model.Person
	err := r.db.WithContext(ctx).Where("stu_id = ?", stuId).First(&person).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // Not found is not an error for us checking existence
	}
	if err != nil {
		return nil, err
	}
	return &person, nil
}

// FindByOpenId 根据OpenID查询
func (r *PersonRepo) FindByOpenId(ctx context.Context, openId string) (*model.Person, error) {
	var person model.Person
	err := r.db.WithContext(ctx).Where("open_id = ?", openId).First(&person).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &person, nil
}

// Create 创建用户
func (r *PersonRepo) Create(ctx context.Context, tx *gorm.DB, person *model.Person) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(person).Error
}

// Update 更新用户
func (r *PersonRepo) Update(ctx context.Context, tx *gorm.DB, person *model.Person) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Save(person).Error
}

// FindByTeamId 查询队伍成员
func (r *PersonRepo) FindByTeamId(ctx context.Context, teamId int64) ([]model.Person, error) {
	var persons []model.Person
	err := r.db.WithContext(ctx).Where("team_id = ?", teamId).Find(&persons).Error
	return persons, err
}

// ResetTeam Info 重置队伍信息
func (r *PersonRepo) ResetTeamInfo(ctx context.Context, tx *gorm.DB, teamId int64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Model(&model.Person{}).Where("team_id = ?", teamId).Updates(map[string]interface{}{"team_id": 0, "status": 0}).Error
}
