package repo

import (
	"context"
	"errors"

	"app/comm"
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

// ResolveTeamStatus 根据成员状态和队伍当前位置推导队伍状态。
func (r *PeopleRepo) ResolveTeamStatus(ctx context.Context, team *model.Team) (string, error) {
	if team == nil {
		return "", nil
	}

	violatedStatus, err := r.resolveViolatedStatus(ctx, team)
	if err != nil {
		return "", err
	}

	members, err := r.FindPeopleByTeamID(ctx, team.ID)
	if err != nil {
		return "", err
	}
	if len(members) == 0 {
		return "", nil
	}

	memberCount := 0
	abandonedCount := 0
	withdrawnCount := 0
	notStartCount := 0
	activeCount := 0
	completedCount := 0

	for _, member := range members {
		if member == nil {
			continue
		}
		memberCount++

		switch member.WalkStatus {
		case comm.WalkStatusAbandoned:
			abandonedCount++
			continue
		case comm.WalkStatusWithdrawn:
			withdrawnCount++
			continue
		case comm.WalkStatusNotStart, comm.WalkStatusPending:
			if member.WalkStatus == comm.WalkStatusPending {
				activeCount++
			} else {
				notStartCount++
			}
		case comm.WalkStatusInProgress:
			activeCount++
		case comm.WalkStatusCompleted:
			completedCount++
		case comm.WalkStatusViolated:
			switch violatedStatus {
			case comm.TeamStatusNotStart:
				notStartCount++
			case comm.TeamStatusCompleted:
				completedCount++
			default:
				activeCount++
			}
		default:
			continue
		}
	}

	if memberCount == 0 {
		return "", nil
	}
	if abandonedCount == memberCount {
		return comm.TeamStatusCompleted, nil
	}
	if abandonedCount+withdrawnCount == memberCount {
		return comm.TeamStatusWithdrawn, nil
	}

	effectiveCount := memberCount - abandonedCount - withdrawnCount
	if activeCount > 0 {
		return comm.TeamStatusInProgress, nil
	}
	if notStartCount == effectiveCount {
		return comm.TeamStatusNotStart, nil
	}
	if completedCount == effectiveCount {
		return comm.TeamStatusCompleted, nil
	}
	return comm.TeamStatusInProgress, nil
}

func (r *PeopleRepo) resolveViolatedStatus(ctx context.Context, team *model.Team) (string, error) {
	if team.PrevPointName == "" {
		return comm.TeamStatusNotStart, nil
	}

	isEndPoint, err := r.isRouteEndPoint(ctx, team)
	if err != nil {
		return "", err
	}
	if isEndPoint {
		return comm.TeamStatusCompleted, nil
	}
	return comm.TeamStatusInProgress, nil
}

func (r *PeopleRepo) isRouteEndPoint(ctx context.Context, team *model.Team) (bool, error) {
	var total int64
	err := r.query.RouteEdge.WithContext(ctx).
		UnderlyingDB().
		Table("route_edges").
		Where(
			"route_name = ? AND point_name = ? AND seq_order = (SELECT MAX(seq_order) FROM route_edges WHERE route_name = ?)",
			team.RouteName,
			team.PrevPointName,
			team.RouteName,
		).
		Count(&total).Error
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *PeopleRepo) UpdateWalkStatus(ctx context.Context, userID int64, status string) error {
	p := r.query.People
	person, err := r.FindPeopleByID(ctx, userID)
	if err != nil {
		return err
	}

	_, err = p.WithContext(ctx).
		Where(p.ID.Eq(userID)).
		Update(p.WalkStatus, status)
	if err != nil {
		return err
	}
	if person != nil {
		_ = peoplecache.DelPersonByOpenID(ctx, person.OpenID)
	}
	return err
}

func (r *PeopleRepo) UpdateTeamIDByUserIDs(ctx context.Context, userIDs []int64, teamID int64) error {
	p := r.query.People
	people, err := r.FindPeopleByIDs(ctx, userIDs)
	if err != nil {
		return err
	}

	_, err = p.WithContext(ctx).
		Where(p.ID.In(userIDs...)).
		Update(p.TeamID, teamID)
	if err != nil {
		return err
	}
	for _, person := range people {
		if person == nil {
			continue
		}
		_ = peoplecache.DelPersonByOpenID(ctx, person.OpenID)
	}
	return err
}

func (r *PeopleRepo) UpdateRoleByUserID(ctx context.Context, userID int64, role string) error {
	p := r.query.People
	person, err := r.FindPeopleByID(ctx, userID)
	if err != nil {
		return err
	}

	_, err = p.WithContext(ctx).
		Where(p.ID.Eq(userID)).
		Update(p.Role, role)
	if err != nil {
		return err
	}
	if person != nil {
		_ = peoplecache.DelPersonByOpenID(ctx, person.OpenID)
	}
	return err
}

func (r *PeopleRepo) UpdateRoleByUserIDs(ctx context.Context, userIDs []int64, role string) error {
	if len(userIDs) == 0 {
		return nil
	}
	p := r.query.People
	people, err := r.FindPeopleByIDs(ctx, userIDs)
	if err != nil {
		return err
	}

	_, err = p.WithContext(ctx).
		Where(p.ID.In(userIDs...)).
		Update(p.Role, role)
	if err != nil {
		return err
	}
	for _, person := range people {
		if person == nil {
			continue
		}
		_ = peoplecache.DelPersonByOpenID(ctx, person.OpenID)
	}
	return err
}

func (r *PeopleRepo) UpdateMembersWalkStatusByCurrent(ctx context.Context, teamID int64, fromStatus string, toStatus string) error {
	p := r.query.People
	people, err := p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(fromStatus),
		).
		Find()
	if err != nil {
		return err
	}

	_, err = p.WithContext(ctx).
		Where(
			p.TeamID.Eq(teamID),
			p.WalkStatus.Eq(fromStatus),
		).
		Update(p.WalkStatus, toStatus)
	if err != nil {
		return err
	}
	for _, person := range people {
		if person == nil {
			continue
		}
		_ = peoplecache.DelPersonByOpenID(ctx, person.OpenID)
	}
	return err
}

func (r *PeopleRepo) UpdateWalkStatusByCurrent(ctx context.Context, fromStatus string, toStatus string) (int64, []int64, error) {
	p := r.query.People
	people, err := p.WithContext(ctx).
		Where(p.WalkStatus.Eq(fromStatus)).
		Find()
	if err != nil {
		return 0, nil, err
	}
	if len(people) == 0 {
		return 0, nil, nil
	}

	_, err = p.WithContext(ctx).
		Where(p.WalkStatus.Eq(fromStatus)).
		Update(p.WalkStatus, toStatus)
	if err != nil {
		return 0, nil, err
	}

	teamIDSet := make(map[int64]struct{})
	for _, person := range people {
		if person == nil {
			continue
		}
		_ = peoplecache.DelPersonByOpenID(ctx, person.OpenID)
		if person.TeamID > 0 {
			teamIDSet[person.TeamID] = struct{}{}
		}
	}

	teamIDs := make([]int64, 0, len(teamIDSet))
	for teamID := range teamIDSet {
		teamIDs = append(teamIDs, teamID)
	}
	return int64(len(people)), teamIDs, nil
}
