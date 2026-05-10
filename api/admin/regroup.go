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

func RegroupHandler() gin.HandlerFunc {
	api := RegroupApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(regroup).Pointer()).Name()] = api
	return regroup
}

type RegroupApi struct {
	Info     struct{} `name:"重组队伍"`
	Request  RegroupApiRequest
	Response RegroupApiResponse
}

type RegroupApiRequest struct {
	Body struct {
		Members   []int  `json:"members" desc:"用户编号,长度3-6人" binding:"required"`
		RouteName string `json:"route_name" desc:"路线名称" binding:"required"`
	}
}

type RegroupApiResponse struct {
	TeamID int `json:"team_id" desc:"队伍编号"`
}

// Run Api业务逻辑执行点
func (r *RegroupApi) Run(ctx *gin.Context) kit.Code {
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

		members, err := txPeopleRepo.FindPeopleByIDs(ctx, memberIDs)
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

		newTeam := &model.Team{
			Name:          "",
			Num:           uint8(len(memberIDs)),
			Password:      "",
			Slogan:        "",
			AllowMatch:    false,
			Captain:       newCaptain.OpenID,
			Submit:        true,
			RouteName:     r.Request.Body.RouteName,
			PrevPointName: "",
			Status:        comm.TeamStatusNotStart,
			IsWrongRoute:  false,
			IsReunite:     true,
			Code:          "",
			Time:          time.Now(),
			IsLost:        false,
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

			if err := txTeamRepo.UpdateByID(ctx, oldTeamID, map[string]any{"num": int8(remainingCount)}); err != nil {
				return err
			}

			inProgressCount, err := txPeopleRepo.CountMembersByStatus(ctx, oldTeamID, comm.WalkStatusInProgress)
			if err != nil {
				return err
			}
			if inProgressCount == 0 {
				if err := txTeamRepo.UpdateByID(ctx, oldTeamID, map[string]any{"status": comm.TeamStatusCompleted}); err != nil {
					return err
				}
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
		nlog.Pick().WithContext(ctx).WithError(err).Error("重组队伍失败")
		return comm.CodeUnknownError
	}

	r.Response.TeamID = int(teamID)
	return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (r *RegroupApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&r.Request.Body)
	if err != nil {
		return err
	}
	return err
}

func regroup(ctx *gin.Context) {
	api := &RegroupApi{}
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
