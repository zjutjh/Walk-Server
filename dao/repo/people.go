package repo

import (
	"context"
	"errors"

	peoplecache "app/dao/cache/people"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"

	"app/dao/model"
	"app/dao/query"
)

type PeopleRepo struct {
	query *query.Query
}

func NewPeopleRepo() *PeopleRepo {
	return &PeopleRepo{
		query: query.Use(ndb.Pick()),
	}
}

func NewPeopleRepoWithDB(db *gorm.DB) *PeopleRepo {
	return &PeopleRepo{
		query: query.Use(db),
	}
}

func NewPeopleRepoWithTx(tx *query.Query) *PeopleRepo {
	return &PeopleRepo{query: tx}
}

// FindByID 根据ID查询人员
func (r *PeopleRepo) FindPeopleByID(ctx context.Context, id int64) (*model.People, error) {
	p := r.query.People
	record, err := p.WithContext(ctx).Where(p.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// FindByOpenID 根据OpenID查询人员
func (r *PeopleRepo) FindPeopleByOpenID(ctx context.Context, openID string) (*model.People, error) {
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

func (r *PeopleRepo) FindPeopleByIdentity(ctx context.Context, identity string) (*model.People, error) {
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

func (r *PeopleRepo) FindPeopleByStuID(ctx context.Context, stuID string) (*model.People, error) {
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

// FindByTeamID 查询队伍成员
func (r *PeopleRepo) FindPeopleByTeamID(ctx context.Context, teamID int64) ([]*model.People, error) {
	p := r.query.People
	return p.WithContext(ctx).
		Where(p.TeamID.Eq(teamID)).
		Order(p.ID).
		Find()
}

func (r *PeopleRepo) ListByTeamID(ctx context.Context, teamID int64) ([]model.People, error) {
	records, err := r.FindPeopleByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	people := make([]model.People, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		people = append(people, *record)
	}
	return people, nil
}

func (r *PeopleRepo) Create(ctx context.Context, person *model.People) error {
	if err := r.query.People.WithContext(ctx).Create(person); err != nil {
		return err
	}
	_ = peoplecache.SetPersonByOpenID(ctx, person)
	return nil
}

func (r *PeopleRepo) UpdateByOpenID(ctx context.Context, openID string, updates map[string]any) error {
	_, err := r.query.People.WithContext(ctx).
		Where(r.query.People.OpenID.Eq(openID)).
		Updates(updates)
	if err != nil {
		return err
	}
	_ = peoplecache.DelPersonByOpenID(ctx, openID)
	return err
}

func (r *PeopleRepo) UpdateByTeamID(ctx context.Context, teamID int64, updates map[string]any) error {
	members, err := r.FindPeopleByTeamID(ctx, teamID)
	if err != nil {
		return err
	}

	_, err = r.query.People.WithContext(ctx).
		Where(r.query.People.TeamID.Eq(teamID)).
		Updates(updates)
	if err != nil {
		return err
	}

	for _, member := range members {
		if member == nil {
			continue
		}
		_ = peoplecache.DelPersonByOpenID(ctx, member.OpenID)
	}
	return nil
}

func (r *PeopleRepo) FindPeopleByIDs(ctx context.Context, ids []int64) ([]*model.People, error) {
	p := r.query.People
	return p.WithContext(ctx).
		Where(p.ID.In(ids...)).
		Find()
}

func (r *PeopleRepo) CountMembersByTeamID(ctx context.Context, teamID int64) (int64, error) {
	p := r.query.People
	return p.WithContext(ctx).
		Where(p.TeamID.Eq(teamID)).
		Count()
}

func (r *PeopleRepo) CountMembersByStatus(ctx context.Context, teamID int64, walkStatus string) (int64, error) {
	p := r.query.People
	return p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(walkStatus),
		).
		Count()
}

func (r *PeopleRepo) UpdateWalkStatus(ctx context.Context, userID int64, status string) error {
	p := r.query.People
	_, err := p.WithContext(ctx).
		Where(p.ID.Eq(userID)).
		Update(p.WalkStatus, status)
	return err
}

func (r *PeopleRepo) UpdateTeamIDByUserIDs(ctx context.Context, userIDs []int64, teamID int64) error {
	p := r.query.People
	_, err := p.WithContext(ctx).
		Where(p.ID.In(userIDs...)).
		Update(p.TeamID, teamID)
	return err
}

func (r *PeopleRepo) UpdateRoleByUserID(ctx context.Context, userID int64, role string) error {
	p := r.query.People
	_, err := p.WithContext(ctx).
		Where(p.ID.Eq(userID)).
		Update(p.Role, role)
	return err
}

func (r *PeopleRepo) UpdateRoleByUserIDs(ctx context.Context, userIDs []int64, role string) error {
	if len(userIDs) == 0 {
		return nil
	}
	p := r.query.People
	_, err := p.WithContext(ctx).
		Where(p.ID.In(userIDs...)).
		Update(p.Role, role)
	return err
}

func (r *PeopleRepo) UpdateMembersWalkStatusByCurrent(ctx context.Context, teamID int64, fromStatus string, toStatus string) error {
	p := r.query.People
	_, err := p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(fromStatus),
		).
		Update(p.WalkStatus, toStatus)
	return err
}
