package api

import (
	"errors"
	"reflect"
	"runtime"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"gorm.io/gorm"

	"app/comm"
	"app/dao/model"
	"app/dao/query"
	repo "app/dao/repo"
)

func RebuildHandler() gin.HandlerFunc {
	api := RebuildApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(rebuild).Pointer()).Name()] = api
	return rebuild
}

type RebuildApi struct {
	Info     struct{} `name:"重组队伍"`
	Request  RebuildApiRequest
	Response RebuildApiResponse
}

type RebuildApiRequest struct {
	Body struct {
		Members   []int  `json:"members" desc:"用户编号,长度3-6人" binding:"required"`
		RouteName string `json:"route_name" desc:"路线名称" binding:"required"`
	}
}

type RebuildApiResponse struct {
	TeamID int `json:"team_id" desc:"队伍编号"`
}

// Run Api业务逻辑执行点
func (r *RebuildApi) Run(ctx *gin.Context) kit.Code {
	if len(r.Request.Body.Members) < 3 {
		return comm.CodeTeamNotEnough
	} else if len(r.Request.Body.Members) > 6 {
		return comm.CodeTeamFull
	}

	memberIDs := make([]int64, 0, len(r.Request.Body.Members))
	for _, memberID := range r.Request.Body.Members {
		memberIDs = append(memberIDs, int64(memberID))
	}

	var teamID int64
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)
		txAdminRepo := repo.NewAdminRepoWithTx(tx)

		members, err := txPeopleRepo.FindPeopleByIDs(ctx, memberIDs)
		if err != nil {
			return err
		}
		if len(members) != len(memberIDs) {
			return gorm.ErrRecordNotFound
		}

		memberMap := make(map[int64]*model.People, len(members))
		for _, member := range members {
			if member == nil {
				return gorm.ErrRecordNotFound
			}
			if member.WalkStatus != comm.WalkStatusNotStart && member.WalkStatus != comm.WalkStatusPending {
				return gorm.ErrInvalidData
			}
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

		newTeam := &model.Team{
			Name:            "",
			Num:             uint8(len(memberIDs)),
			Password:        "",
			Slogan:          "",
			AllowMatch:      false,
			Captain:         newCaptain.OpenID,
			Submit:          true,
			RouteName:       r.Request.Body.RouteName,
			LatestPointName: "",
			Status:          comm.TeamStatusNotStart,
			IsWrongRoute:    false,
			IsReunite:       true,
			Code:            "",
			Time:            time.Now(),
			IsLost:          false,
		}
		if err := txTeamRepo.Create(ctx, newTeam); err != nil {
			return err
		}

		if err := txPeopleRepo.UpdateTeamIDByUserIDs(ctx, memberIDs, newTeam.ID); err != nil {
			return err
		}
		if err := txPeopleRepo.UpdateRoleByUserIDs(ctx, memberIDs, comm.RoleMember); err != nil {
			return err
		}
		if err := txPeopleRepo.UpdateRoleByUserID(ctx, newCaptain.ID, comm.RoleCaptain); err != nil {
			return err
		}
		if err := r.handleStartPointCheckin(ctx, txTeamRepo, txPeopleRepo, txAdminRepo, newTeam); err != nil {
			return err
		}

		for _, oldTeamID := range oldTeamIDs {
			remainingCount, err := txPeopleRepo.CountMembersByTeamID(ctx, oldTeamID)
			if err != nil {
				return err
			}
			if remainingCount == 0 {
				if err := txTeamRepo.DeleteByID(ctx, oldTeamID); err != nil {
					return err
				}
				continue
			}

			oldTeam, err := txTeamRepo.FindTeamByID(ctx, oldTeamID)
			if err != nil {
				return err
			}
			if oldTeam == nil {
				return gorm.ErrRecordNotFound
			}

			nextStatus, err := txPeopleRepo.ResolveTeamStatus(ctx, oldTeam)
			if err != nil {
				return err
			}
			updates := map[string]any{"num": int8(remainingCount)}
			if nextStatus != "" {
				updates["status"] = nextStatus
			}
			if err := txTeamRepo.UpdateByID(ctx, oldTeamID, updates); err != nil {
				return err
			}

			remainingMembers, err := txPeopleRepo.FindPeopleByTeamID(ctx, oldTeamID)
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
				if err := txTeamRepo.UpdateByID(ctx, oldTeamID, map[string]any{"captain": nextCaptain.OpenID}); err != nil {
					return err
				}
				if err := txPeopleRepo.UpdateRoleByUserID(ctx, nextCaptain.ID, comm.RoleCaptain); err != nil {
					return err
				}
			}
		}

		teamID = newTeam.ID
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return comm.CodeDataNotFound
		}
		if errors.Is(err, gorm.ErrInvalidData) {
			return comm.CodeParameterInvalid
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("重组队伍失败")
		return comm.CodeServerError
	}

	r.Response.TeamID = int(teamID)
	return comm.CodeOK
}

func (r *RebuildApi) handleStartPointCheckin(ctx *gin.Context, teamRepo *repo.TeamRepo, peopleRepo *repo.PeopleRepo, adminRepo *repo.AdminRepo, team *model.Team) error {
	startEdge, err := teamRepo.FindRouteStartEdge(ctx, team.RouteName)
	if err != nil {
		return err
	}
	if startEdge == nil {
		return gorm.ErrRecordNotFound
	}

	admin, err := adminRepo.FindByPointName(ctx, startEdge.PointName)
	if err != nil {
		return err
	}
	if admin == nil {
		return gorm.ErrRecordNotFound
	}

	if err := teamRepo.UpdateLatestPointName(ctx, team.ID, startEdge.PointName); err != nil {
		return err
	}
	if err := teamRepo.CreateCheckin(ctx, admin.ID, team.ID, startEdge.PointName, team.RouteName); err != nil {
		return err
	}
	return peopleRepo.UpdateMembersWalkStatusByCurrent(ctx, team.ID, comm.WalkStatusNotStart, comm.WalkStatusPending)
}

// Run Api初始化 进行参数校验和绑定
func (r *RebuildApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&r.Request.Body)
	if err != nil {
		return err
	}
	return err
}

func rebuild(ctx *gin.Context) {
	api := &RebuildApi{}
	err := api.Init(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
