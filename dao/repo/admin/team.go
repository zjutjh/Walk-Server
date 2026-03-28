package repo

import (
	routeCache "app/dao/cache/route"
	teamCache "app/dao/cache/team"
	"context"
	"errors"
	"slices"
	"time"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"

	"app/comm"
	"app/dao/model"
	"app/dao/query"
)

type TeamRepo struct {
	query *query.Query
}

func NewTeamRepo() *TeamRepo {
	return &TeamRepo{
		query: query.Use(ndb.Pick()),
	}

}

// FindTeamByID 根据ID查询队伍
func (r *TeamRepo) FindTeamByID(ctx context.Context, id int64) (*model.Team, error) {
	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if record.Code != "" {
		_ = teamCache.SetTeamIDByCode(ctx, record.Code, record.ID)
	}
	return record, nil
}

func (r *TeamRepo) FindByCode(ctx context.Context, code string) (*model.Team, error) {
	if teamID, hit, err := teamCache.GetTeamIDByCode(ctx, code); err == nil && hit {
		return r.FindTeamByID(ctx, teamID)
	}

	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.Code.Eq(code)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = teamCache.SetTeamIDByCode(ctx, record.Code, record.ID)
	return record, nil
}

func (r *TeamRepo) findTeamByID(ctx context.Context, tx *query.Query, id int64) (*model.Team, error) {
	t := tx.Team
	record, err := t.WithContext(ctx).Where(t.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *TeamRepo) createRegroupTeam(ctx context.Context, tx *query.Query, memberCount int, routeName string, captainOpenID string) (*model.Team, error) {
	t := tx.Team
	team := &model.Team{
		Name:          "",
		Num:           int8(memberCount),
		Password:      "",
		Slogan:        "",
		AllowMatch:    0,
		Captain:       captainOpenID,
		Submit:        1,
		RouteName:     routeName,
		PrevPointName: "",
		Status:        comm.TeamStatusNotStart,
		IsWrongRoute:  0,
		IsReunite:     1,
		Code:          "",
		Time:          time.Now(),
		IsLost:        0,
	}
	if err := t.WithContext(ctx).Create(team); err != nil {
		return nil, err
	}
	return team, nil
}

func (r *TeamRepo) deleteTeams(ctx context.Context, tx *query.Query, teamIDs []int64) error {
	if len(teamIDs) == 0 {
		return nil
	}
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.In(teamIDs...)).
		Delete()
	return err
}

func (r *TeamRepo) updateTeamCaptain(ctx context.Context, tx *query.Query, teamID int64, captainOpenID string) error {
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.Captain, captainOpenID)
	return err
}

func (r *TeamRepo) updateTeamNum(ctx context.Context, tx *query.Query, teamID int64, memberCount int64) error {
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.Num, int8(memberCount))
	return err
}

func (r *TeamRepo) bindCode(ctx context.Context, tx *query.Query, teamID int64, content string) error {
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.Code, content)
	return err
}

func (r *TeamRepo) updateTeamStatus(ctx context.Context, tx *query.Query, teamID int64, status string) error {
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.Status, status)
	return err
}

func (r *TeamRepo) CreateCheckin(ctx context.Context, adminID, teamID int64, pointName, routeName string) error {
	checkin := &model.Checkin{
		AdminID:   adminID,
		TeamID:    teamID,
		PointName: pointName,
		RouteName: routeName,
		Time:      time.Now(),
	}
	return r.query.Checkin.WithContext(ctx).Create(checkin)
}

func (r *TeamRepo) UpdateTeamWrongRoute(ctx context.Context, teamID int64, isWrongRoute int8) error {
	return r.query.Transaction(func(tx *query.Query) error {
		t := tx.Team
		_, err := t.WithContext(ctx).
			Where(t.ID.Eq(teamID)).
			Update(t.IsWrongRoute, isWrongRoute)
		return err
	})
}

func (r *TeamRepo) CreateWrongRouteRecord(ctx context.Context, teamID int64, routeName, wrongRouteName string, adminID int64) error {
	record := &model.WrongRouteRecord{
		TeamID:         teamID,
		RouteName:      routeName,
		WrongRouteName: wrongRouteName,
		AdminID:        adminID,
	}
	return r.query.WrongRouteRecord.WithContext(ctx).Create(record)
}

func (r *TeamRepo) ClearLostStatus(ctx context.Context, teamID int64) error {
	return r.query.Transaction(func(tx *query.Query) error {
		t := tx.Team
		_, err := t.WithContext(ctx).
			Where(
				t.ID.Eq(teamID),
				t.IsLost.Eq(1),
			).
			Update(t.IsLost, 0)
		return err
	})
}

func (r *TeamRepo) FindRouteByName(ctx context.Context, routeName string) (*model.Route, error) {
	if route, hit, err := routeCache.GetRoute(ctx, routeName); err == nil && hit {
		return route, nil
	}

	rt := r.query.Route
	record, err := rt.WithContext(ctx).Where(rt.Name.Eq(routeName)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = routeCache.SetRoute(ctx, record)
	return record, nil
}

func (r *TeamRepo) FindRouteEdge(ctx context.Context, routeName, pointName string) (*model.RouteEdge, error) {
	if routeEdge, hit, err := routeCache.GetRouteEdge(ctx, routeName, pointName); err == nil && hit {
		return routeEdge, nil
	}

	re := r.query.RouteEdge
	record, err := re.WithContext(ctx).
		Where(
			re.RouteName.Eq(routeName),
			re.PointName.Eq(pointName),
		).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = routeCache.SetRouteEdge(ctx, record)
	return record, nil
}

func (r *TeamRepo) FindPointRoutes(ctx context.Context, pointName string) ([]string, error) {
	if routeNames, hit, err := routeCache.GetPointRoutes(ctx, pointName); err == nil && hit {
		return routeNames, nil
	}

	re := r.query.RouteEdge
	var routeNames []string
	err := re.WithContext(ctx).
		Where(re.PointName.Eq(pointName)).
		Pluck(re.RouteName, &routeNames)
	if err != nil {
		return nil, err
	}
	_ = routeCache.SetPointRoutes(ctx, pointName, routeNames)
	return routeNames, nil
}

func (r *TeamRepo) StartPointCheckin(ctx context.Context, teamID int64, pointName string) error {
	peopleRepo := NewPeopleRepo()

	return r.query.Transaction(func(tx *query.Query) error {
		t := tx.Team
		_, err := t.WithContext(ctx).
			Where(t.ID.Eq(teamID)).
			Update(t.PrevPointName, pointName)
		if err != nil {
			return err
		}
		return peopleRepo.setAllMembersPending(ctx, tx, teamID)
	})
}

func (r *TeamRepo) UpdatePrevPointName(ctx context.Context, teamID int64, pointName string) error {
	return r.query.Transaction(func(tx *query.Query) error {
		t := tx.Team
		_, err := t.WithContext(ctx).
			Where(t.ID.Eq(teamID)).
			Update(t.PrevPointName, pointName)
		if err != nil {
			return err
		}
		return nil
	})
}

// BindCodeAndStartPendingMembers 绑定签到码，并将待出发成员更新为进行中
func (r *TeamRepo) BindCodeAndStartPendingMembers(ctx context.Context, teamID int64, content string) error {
	peopleRepo := NewPeopleRepo()

	err := r.query.Transaction(func(tx *query.Query) error {
		if err := r.bindCode(ctx, tx, teamID, content); err != nil {
			return err
		}
		if err := peopleRepo.updateMembersWalkStatusByCurrent(ctx, tx, teamID, comm.WalkStatusPending, comm.WalkStatusInProgress); err != nil {
			return err
		}
		inProgressCount, err := peopleRepo.countInProgressMembers(ctx, tx, teamID)
		if err != nil {
			return err
		}
		if inProgressCount > 0 {
			err = r.updateTeamStatus(ctx, tx, teamID, comm.TeamStatusInProgress)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	_ = teamCache.SetTeamIDByCode(ctx, content, teamID)
	return nil
}

// ConfirmDestination 将队伍和队伍成员状态更新为 completed
func (r *TeamRepo) ConfirmDestination(ctx context.Context, teamID int64) error {
	peopleRepo := NewPeopleRepo()

	return r.query.Transaction(func(tx *query.Query) error {
		if err := peopleRepo.updateMembersWalkStatusByCurrent(ctx, tx, teamID, comm.WalkStatusInProgress, comm.WalkStatusCompleted); err != nil {
			return err
		}
		return r.updateTeamStatus(ctx, tx, teamID, comm.TeamStatusCompleted)
	})
}

// MarkViolation 将队伍状态更新为 completed，并将进行中的成员更新为 violated
func (r *TeamRepo) MarkViolation(ctx context.Context, teamID int64) error {
	peopleRepo := NewPeopleRepo()

	return r.query.Transaction(func(tx *query.Query) error {
		if err := r.updateTeamStatus(ctx, tx, teamID, comm.TeamStatusCompleted); err != nil {
			return err
		}
		return peopleRepo.updateMembersWalkStatusByCurrent(ctx, tx, teamID, comm.WalkStatusInProgress, comm.WalkStatusViolated)
	})
}

// UpdateUserStatus 更改人员状态，并根据队伍成员状态回推队伍状态
func (r *TeamRepo) UpdateUserStatus(ctx context.Context, user *model.People, status string) error {
	peopleRepo := NewPeopleRepo()

	return r.query.Transaction(func(tx *query.Query) error {
		if err := peopleRepo.updateWalkStatus(ctx, tx, user.ID, status); err != nil {
			return err
		}
		team, err := r.findTeamByID(ctx, tx, user.TeamID)
		if err != nil {
			return err
		}
		if team == nil {
			return gorm.ErrRecordNotFound
		}

		if team.Status == comm.TeamStatusNotStart {
			memberCount, err := peopleRepo.countMembersByTeamID(ctx, tx, user.TeamID)
			if err != nil {
				return err
			}
			abandonedCount, err := peopleRepo.countMembersByStatus(ctx, tx, user.TeamID, comm.WalkStatusAbandoned)
			if err != nil {
				return err
			}
			if memberCount > 0 && memberCount == abandonedCount {
				return r.updateTeamStatus(ctx, tx, user.TeamID, comm.TeamStatusCompleted)
			}
			return nil
		}

		inProgressCount, err := peopleRepo.countInProgressMembers(ctx, tx, user.TeamID)
		if err != nil {
			return err
		}
		if inProgressCount > 0 {
			if team.Status != comm.TeamStatusInProgress {
				return r.updateTeamStatus(ctx, tx, user.TeamID, comm.TeamStatusInProgress)
			}
			return nil
		}

		if status != comm.WalkStatusWithdrawn {
			return r.updateTeamStatus(ctx, tx, user.TeamID, comm.TeamStatusCompleted)
		}

		if team != nil && team.Status != comm.TeamStatusCompleted {
			return r.updateTeamStatus(ctx, tx, user.TeamID, comm.TeamStatusWithDrawn)
		}

		return nil
	})
}

// Regroup 创建新队伍
func (r *TeamRepo) Regroup(ctx context.Context, memberIDs []int64, routeName string) (int64, error) {
	peopleRepo := NewPeopleRepo()

	var newTeamID int64
	err := r.query.Transaction(func(tx *query.Query) error {
		members, err := peopleRepo.findPeopleByIDs(ctx, tx, memberIDs)
		if err != nil {
			return err
		}
		if len(members) != len(memberIDs) {
			return gorm.ErrRecordNotFound
		}

		memberMap := make(map[int64]*model.People, len(members))
		for _, member := range members {
			memberMap[member.ID] = member
		}

		newCaptain, ok := memberMap[memberIDs[0]]
		if !ok {
			return gorm.ErrRecordNotFound
		}

		oldTeamIDs := make([]int64, 0, len(members))
		for _, member := range members {
			if member.TeamID > 0 {
				oldTeamIDs = append(oldTeamIDs, member.TeamID)
			}
		}
		slices.Sort(oldTeamIDs)
		oldTeamIDs = slices.Compact(oldTeamIDs)

		newTeam, err := r.createRegroupTeam(ctx, tx, len(memberIDs), routeName, newCaptain.OpenID)
		if err != nil {
			return err
		}

		if err := peopleRepo.updateTeamIDByUserIDs(ctx, tx, memberIDs, newTeam.ID); err != nil {
			return err
		}
		if err := peopleRepo.updateRoleByUserIDs(ctx, tx, memberIDs, comm.RoleMember); err != nil {
			return err
		}
		if err := peopleRepo.updateRoleByUserID(ctx, tx, newCaptain.ID, comm.RoleCaptain); err != nil {
			return err
		}

		var deleteTeamIDs []int64
		for _, oldTeamID := range oldTeamIDs {
			remainingCount, err := peopleRepo.countMembersByTeamID(ctx, tx, oldTeamID)
			if err != nil {
				return err
			}
			if remainingCount == 0 {
				deleteTeamIDs = append(deleteTeamIDs, oldTeamID)
				continue
			}
			if err := r.updateTeamNum(ctx, tx, oldTeamID, remainingCount); err != nil {
				return err
			}

			inProgressCount, err := peopleRepo.countInProgressMembers(ctx, tx, oldTeamID)
			if err != nil {
				return err
			}
			if inProgressCount == 0 {
				if err := r.updateTeamStatus(ctx, tx, oldTeamID, comm.TeamStatusCompleted); err != nil {
					return err
				}
			}

			remainingMembers, err := peopleRepo.findPeopleByTeamID(ctx, tx, oldTeamID)
			if err != nil {
				return err
			}

			captainStillExists := false
			var nextCaptain *model.People
			for _, member := range remainingMembers {
				if member.Role == comm.RoleCaptain {
					captainStillExists = true
				}
				if nextCaptain == nil {
					nextCaptain = member
				}
			}

			if !captainStillExists && nextCaptain != nil {
				if err := r.updateTeamCaptain(ctx, tx, oldTeamID, nextCaptain.OpenID); err != nil {
					return err
				}
				if err := peopleRepo.updateRoleByUserID(ctx, tx, nextCaptain.ID, comm.RoleCaptain); err != nil {
					return err
				}
			}
		}

		if err := r.deleteTeams(ctx, tx, deleteTeamIDs); err != nil {
			return err
		}

		newTeamID = newTeam.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return newTeamID, nil
}
