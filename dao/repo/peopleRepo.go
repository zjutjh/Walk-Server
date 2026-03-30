package repo

import (
	"context"
	"errors"

	peoplecache "app/dao/cache/people"
	"app/dao/model"
	"app/dao/query"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type PeopleRepo struct {
	query *query.Query
}

func NewPeopleRepo() *PeopleRepo {
	return &PeopleRepo{query: query.Use(ndb.Pick())}
}

func NewPeopleRepoWithDB(db *gorm.DB) *PeopleRepo {
	return &PeopleRepo{query: query.Use(db)}
}

func NewPeopleRepoWithTx(tx *query.Query) *PeopleRepo {
	return &PeopleRepo{query: tx}
}

func (r *PeopleRepo) FindByOpenID(ctx context.Context, openID string) (*model.People, error) {
	if people, hit, err := peoplecache.GetPersonByOpenID(ctx, openID); err == nil && hit {
		return people, nil
	}

	p := r.query.People
	record, err := p.WithContext(ctx).Where(p.OpenID.Eq(openID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = peoplecache.SetPersonByOpenID(ctx, record)
	return record, nil
}

func (r *PeopleRepo) FindByIdentity(ctx context.Context, identity string) (*model.People, error) {
	p := r.query.People
	record, err := p.WithContext(ctx).Where(p.Identity.Eq(identity)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *PeopleRepo) FindByStuID(ctx context.Context, stuID string) (*model.People, error) {
	if stuID == "" {
		return nil, nil
	}
	p := r.query.People
	record, err := p.WithContext(ctx).Where(p.StuID.Eq(stuID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *PeopleRepo) Create(ctx context.Context, person *model.People) error {
	p := r.query.People
	if err := p.WithContext(ctx).Create(person); err != nil {
		return err
	}
	_ = peoplecache.SetPersonByOpenID(ctx, person)
	return nil
}

func (r *PeopleRepo) UpdateByOpenID(ctx context.Context, openID string, updates map[string]any) error {
	p := r.query.People
	_, err := p.WithContext(ctx).Where(p.OpenID.Eq(openID)).Updates(updates)
	if err != nil {
		return err
	}
	_ = peoplecache.DelPersonByOpenID(ctx, openID)
	return nil
}

func (r *PeopleRepo) UpdateByTeamID(ctx context.Context, teamID int64, updates map[string]any) error {
	members, err := r.ListByTeamID(ctx, teamID)
	if err != nil {
		return err
	}

	p := r.query.People
	_, err = p.WithContext(ctx).Where(p.TeamID.Eq(teamID)).Updates(updates)
	if err != nil {
		return err
	}

	for _, member := range members {
		_ = peoplecache.DelPersonByOpenID(ctx, member.OpenID)
	}
	return nil
}

func (r *PeopleRepo) ListByTeamID(ctx context.Context, teamID int64) ([]*model.People, error) {
	p := r.query.People
	return p.WithContext(ctx).
		Where(p.TeamID.Eq(teamID)).
		Order(p.Role.Desc(), p.ID).
		Find()
}
