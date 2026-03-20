package repo

import (
	"context"
	"errors"
	"slices"

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

// FindByID 根据ID查询队伍
func (r *TeamRepo) FindByID(ctx context.Context, id int64) (*model.Team, error) {
	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

// FindByCode 根据签到码查询队伍
func (r *TeamRepo) FindByCode(ctx context.Context, code string) (*model.Team, error) {
	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.Code.Eq(code)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *TeamRepo) findByID(ctx context.Context, tx *query.Query, id int64) (*model.Team, error) {
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

func (r *TeamRepo) createRegroupTeam(ctx context.Context, tx *query.Query, memberCount int, routeName string) (*model.Team, error) {
	t := tx.Team
	team := &model.Team{
		Name:          "",
		Num:           int8(memberCount),
		Password:      "",
		Slogan:        "",
		AllowMatch:    0,
		Captain:       "",
		Submit:        1,
		RouteName:     routeName,
		PrevPointName: "",
		Status:        comm.TeamStatusNotStart,
		IsWrongRoute:  0,
		IDReunite:     1,
		Code:          "",
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

func (r *TeamRepo) updateCaptain(ctx context.Context, tx *query.Query, teamID int64, captainOpenID string) error {
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.Captain, captainOpenID)
	return err
}

func (r *TeamRepo) bindCode(ctx context.Context, tx *query.Query, teamID int64, content string) error {
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.Code, content)
	return err
}

func (r *TeamRepo) updateStatusByInProgressCount(ctx context.Context, tx *query.Query, teamID int64, inProgressCount int64) error {
	t := tx.Team
	status := comm.TeamStatusCompleted
	if inProgressCount > 0 {
		status = comm.TeamStatusInProgress
	}

	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.Status, status)
	return err
}

func (r *TeamRepo) updateStatus(ctx context.Context, tx *query.Query, teamID int64, status string) error {
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.Status, status)
	return err
}

func (r *TeamRepo) completeTeam(ctx context.Context, tx *query.Query, teamID int64) error {
	return r.updateStatus(ctx, tx, teamID, comm.TeamStatusCompleted)
}

func (r *TeamRepo) completeTeams(ctx context.Context, tx *query.Query, teamIDs []int64) error {
	if len(teamIDs) == 0 {
		return nil
	}
	t := tx.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.In(teamIDs...)).
		Update(t.Status, comm.TeamStatusCompleted)
	return err
}

// BindCodeAndStartPendingMembers 绑定签到码，并将待出发成员更新为进行中
func (r *TeamRepo) BindCodeAndStartPendingMembers(ctx context.Context, teamID int64, content string) error {
	peopleRepo := NewPeopleRepo()

	return r.query.Transaction(func(tx *query.Query) error {
		if err := r.bindCode(ctx, tx, teamID, content); err != nil {
			return err
		}
		if err := peopleRepo.startPendingMembers(ctx, tx, teamID); err != nil {
			return err
		}
		inProgressCount, err := peopleRepo.countInProgressMembers(ctx, tx, teamID)
		if err != nil {
			return err
		}
		return r.updateStatusByInProgressCount(ctx, tx, teamID, inProgressCount)
	})
}

// ConfirmDestination 将队伍和队伍成员状态更新为 completed
func (r *TeamRepo) ConfirmDestination(ctx context.Context, teamID int64) error {
	peopleRepo := NewPeopleRepo()

	return r.query.Transaction(func(tx *query.Query) error {
		if err := peopleRepo.completeAllMembers(ctx, tx, teamID); err != nil {
			return err
		}
		return r.completeTeam(ctx, tx, teamID)
	})
}

// MarkViolation 将队伍状态更新为 completed，并将进行中的成员更新为 violated
func (r *TeamRepo) MarkViolation(ctx context.Context, teamID int64) error {
	peopleRepo := NewPeopleRepo()

	return r.query.Transaction(func(tx *query.Query) error {
		if err := r.completeTeam(ctx, tx, teamID); err != nil {
			return err
		}
		return peopleRepo.violateInProgressMembers(ctx, tx, teamID)
	})
}

// UpdateUserStatus 更改人员状态，并根据队伍成员状态回推队伍状态
func (r *TeamRepo) UpdateUserStatus(ctx context.Context, userID int64, status string) error {
	peopleRepo := NewPeopleRepo()

	return r.query.Transaction(func(tx *query.Query) error {
		user, err := peopleRepo.findByID(ctx, tx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			return gorm.ErrRecordNotFound
		}

		if err := peopleRepo.updateWalkStatus(ctx, tx, userID, status); err != nil {
			return err
		}

		if user.TeamID <= 0 {
			return nil
		}

		inProgressCount, err := peopleRepo.countInProgressMembers(ctx, tx, user.TeamID)
		if err != nil {
			return err
		}
		if inProgressCount > 0 {
			return nil
		}

		if status != comm.WalkStatusWithdrawn {
			return r.updateStatus(ctx, tx, user.TeamID, comm.TeamStatusCompleted)
		}

		team := tx.Team
		teamInfo, err := team.WithContext(ctx).
			Where(team.ID.Eq(user.TeamID)).
			First()
		if err != nil {
			return err
		}
		if teamInfo.Status != comm.TeamStatusCompleted {
			return r.updateStatus(ctx, tx, user.TeamID, comm.TeamStatusWithDrawn)
		}

		return nil
	})
}

// Regroup 创建新队伍，将成员原队伍状态改成 completed，并更新成员 team_id
func (r *TeamRepo) Regroup(ctx context.Context, memberIDs []int64, routeName string) (int64, error) {
	peopleRepo := NewPeopleRepo()

	returnTeamID := int64(0)
	err := r.query.Transaction(func(tx *query.Query) error {
		members, err := peopleRepo.findByIDs(ctx, tx, memberIDs)
		if err != nil {
			return err
		}
		if len(members) != len(memberIDs) {
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

		newTeam, err := r.createRegroupTeam(ctx, tx, len(memberIDs), routeName)
		if err != nil {
			return err
		}

		if err := peopleRepo.updateTeamIDByUserIDs(ctx, tx, memberIDs, newTeam.ID); err != nil {
			return err
		}
		if err := peopleRepo.updateRoleByUserIDs(ctx, tx, memberIDs, comm.RoleMember); err != nil {
			return err
		}

		deleteTeamIDs := make([]int64, 0)
		for _, oldTeamID := range oldTeamIDs {
			remainingCount, err := peopleRepo.countByTeamID(ctx, tx, oldTeamID)
			if err != nil {
				return err
			}
			if remainingCount == 0 {
				deleteTeamIDs = append(deleteTeamIDs, oldTeamID)
				continue
			}

			oldTeam, err := r.findByID(ctx, tx, oldTeamID)
			if err != nil {
				return err
			}
			if oldTeam == nil {
				continue
			}

			remainingMembers, err := peopleRepo.findByTeamID(ctx, tx, oldTeamID)
			if err != nil {
				return err
			}

			captainStillExists := false
			var nextCaptain *model.People
			for _, remainingMember := range remainingMembers {
				if remainingMember.OpenID == oldTeam.Captain {
					captainStillExists = true
				}
				if nextCaptain == nil {
					nextCaptain = remainingMember
				}
			}

			if !captainStillExists && nextCaptain != nil {
				if err := r.updateCaptain(ctx, tx, oldTeamID, nextCaptain.OpenID); err != nil {
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

		returnTeamID = newTeam.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return returnTeamID, nil
}
