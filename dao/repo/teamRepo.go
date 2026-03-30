package repo

import (
	"context"
	"errors"

	teamcache "app/dao/cache/team"
	"app/dao/model"
	"app/dao/query"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TeamRepo struct {
	query *query.Query
}

func NewTeamRepo() *TeamRepo {
	return &TeamRepo{query: query.Use(ndb.Pick())}
}

func NewTeamRepoWithDB(db *gorm.DB) *TeamRepo {
	return &TeamRepo{query: query.Use(db)}
}

func NewTeamRepoWithTx(tx *query.Query) *TeamRepo {
	return &TeamRepo{query: tx}
}

func (r *TeamRepo) Create(ctx context.Context, team *model.Team) error {
	t := r.query.Team
	if err := t.WithContext(ctx).Create(team); err != nil {
		return err
	}
	_ = teamcache.SetTeamByID(ctx, team)
	if team.Code != "" {
		_ = teamcache.SetTeamIDByCode(ctx, team.Code, team.ID)
	}
	return nil
}

func (r *TeamRepo) FindByID(ctx context.Context, id int64) (*model.Team, error) {
	if team, hit, err := teamcache.GetTeamByID(ctx, id); err == nil && hit {
		return team, nil
	}

	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = teamcache.SetTeamByID(ctx, record)
	if record.Code != "" {
		_ = teamcache.SetTeamIDByCode(ctx, record.Code, record.ID)
	}
	return record, nil
}

func (r *TeamRepo) FindByName(ctx context.Context, name string) (*model.Team, error) {
	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.Name.Eq(name)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *TeamRepo) FindByNameExceptID(ctx context.Context, name string, id int64) (*model.Team, error) {
	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.Name.Eq(name), t.ID.Neq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *TeamRepo) UpdateByID(ctx context.Context, id int64, updates map[string]any) error {
	t := r.query.Team
	_, err := t.WithContext(ctx).Where(t.ID.Eq(id)).Updates(updates)
	if err != nil {
		return err
	}
	_ = teamcache.DelTeamByID(ctx, id)
	return nil
}

func (r *TeamRepo) IncrementNumIfAvailable(ctx context.Context, id int64, maxTeamSize int) (bool, error) {
	result := r.query.Team.WithContext(ctx).UnderlyingDB().WithContext(ctx).
		Model(&model.Team{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND submit = ? AND num < ?", id, 0, maxTeamSize).
		UpdateColumn("num", gorm.Expr("num + ?", 1))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		_ = teamcache.DelTeamByID(ctx, id)
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DecrementNumIfPositive(ctx context.Context, id int64) (bool, error) {
	result := r.query.Team.WithContext(ctx).UnderlyingDB().WithContext(ctx).
		Model(&model.Team{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND submit = ? AND num > 0", id, 0).
		UpdateColumn("num", gorm.Expr("num - ?", 1))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		_ = teamcache.DelTeamByID(ctx, id)
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DeleteByID(ctx context.Context, id int64) error {
	t := r.query.Team
	_, err := t.WithContext(ctx).Where(t.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	_ = teamcache.DelTeamByID(ctx, id)
	return nil
}
