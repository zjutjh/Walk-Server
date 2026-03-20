package repo

import (
	"context"
	"errors"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"

	"app/comm"
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

// FindByID 根据ID查询人员
func (r *PeopleRepo) FindByID(ctx context.Context, id int64) (*model.People, error) {
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
func (r *PeopleRepo) FindByOpenID(ctx context.Context, openID string) (*model.People, error) {
	p := r.query.People
	record, err := p.WithContext(ctx).Where(p.OpenID.Eq(openID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// FindByTeamID 查询队伍成员
func (r *PeopleRepo) FindByTeamID(ctx context.Context, teamID int64) ([]*model.People, error) {
	p := r.query.People
	return p.WithContext(ctx).
		Where(p.TeamID.Eq(teamID)).
		Order(p.ID).
		Find()
}

func (r *PeopleRepo) findByTeamID(ctx context.Context, tx *query.Query, teamID int64) ([]*model.People, error) {
	p := tx.People
	return p.WithContext(ctx).
		Where(p.TeamID.Eq(teamID)).
		Order(p.ID).
		Find()
}

func (r *PeopleRepo) countByTeamID(ctx context.Context, tx *query.Query, teamID int64) (int64, error) {
	p := tx.People
	return p.WithContext(ctx).
		Where(p.TeamID.Eq(teamID)).
		Count()
}

func (r *PeopleRepo) findByIDs(ctx context.Context, tx *query.Query, ids []int64) ([]*model.People, error) {
	p := tx.People
	return p.WithContext(ctx).
		Where(p.ID.In(ids...)).
		Find()
}

func (r *PeopleRepo) findByID(ctx context.Context, tx *query.Query, id int64) (*model.People, error) {
	p := tx.People
	record, err := p.WithContext(ctx).Where(p.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *PeopleRepo) updateWalkStatus(ctx context.Context, tx *query.Query, userID int64, status string) error {
	p := tx.People
	_, err := p.WithContext(ctx).
		Where(p.ID.Eq(userID)).
		Update(p.WalkStatus, status)
	return err
}

func (r *PeopleRepo) updateTeamIDByUserIDs(ctx context.Context, tx *query.Query, userIDs []int64, teamID int64) error {
	p := tx.People
	_, err := p.WithContext(ctx).
		Where(p.ID.In(userIDs...)).
		Update(p.TeamID, teamID)
	return err
}

func (r *PeopleRepo) updateRoleByUserID(ctx context.Context, tx *query.Query, userID int64, role string) error {
	p := tx.People
	_, err := p.WithContext(ctx).
		Where(p.ID.Eq(userID)).
		Update(p.Role, role)
	return err
}

func (r *PeopleRepo) updateRoleByUserIDs(ctx context.Context, tx *query.Query, userIDs []int64, role string) error {
	if len(userIDs) == 0 {
		return nil
	}
	p := tx.People
	_, err := p.WithContext(ctx).
		Where(p.ID.In(userIDs...)).
		Update(p.Role, role)
	return err
}

func (r *PeopleRepo) startPendingMembers(ctx context.Context, tx *query.Query, teamID int64) error {
	p := tx.People
	_, err := p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(comm.WalkStatusPending),
		).
		Update(p.WalkStatus, comm.WalkStatusInProgress)
	return err
}

func (r *PeopleRepo) countInProgressMembers(ctx context.Context, tx *query.Query, teamID int64) (int64, error) {
	p := tx.People
	return p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(comm.WalkStatusInProgress),
		).
		Count()
}

func (r *PeopleRepo) countCompletedMembers(ctx context.Context, tx *query.Query, teamID int64) (int64, error) {
	p := tx.People
	return p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(comm.WalkStatusCompleted),
		).
		Count()
}

func (r *PeopleRepo) countWithdrawnMembers(ctx context.Context, tx *query.Query, teamID int64) (int64, error) {
	p := tx.People
	return p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(comm.WalkStatusWithdrawn),
		).
		Count()
}

func (r *PeopleRepo) completeAllMembers(ctx context.Context, tx *query.Query, teamID int64) error {
	p := tx.People

	_, err := p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Neq(comm.WalkStatusCompleted),
		).
		Update(p.WalkStatus, comm.WalkStatusCompleted)

	return err
}

func (r *PeopleRepo) violateInProgressMembers(ctx context.Context, tx *query.Query, teamID int64) error {
	p := tx.People

	_, err := p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(comm.WalkStatusInProgress),
		).
		Update(p.WalkStatus, comm.WalkStatusViolated)

	return err
}

// ConfirmDestination 将指定队伍下所有未完成的人员状态更新为 completed
func (r *PeopleRepo) ConfirmDestination(ctx context.Context, teamID int64) error {
	p := r.query.People

	_, err := p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Neq(comm.WalkStatusCompleted),
		).
		Update(p.WalkStatus, comm.WalkStatusCompleted)

	return err
}
