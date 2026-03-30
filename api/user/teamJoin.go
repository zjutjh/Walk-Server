package api

import (
	"errors"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/query"
	"app/dao/repo"
)

func TeamJoinHandler() gin.HandlerFunc {
	api := TeamJoinApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamJoin).Pointer()).Name()] = api
	return hfTeamJoin
}

type TeamJoinApi struct {
	Info     struct{} `name:"加入队伍" desc:"加入已有队伍"`
	Request  TeamJoinApiRequest
	Response struct{}
}

type TeamJoinApiRequest struct {
	TeamID   int64  `json:"team_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *TeamJoinApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *TeamJoinApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindPeopleByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil {
		return comm.CodeDataNotFound
	}
	if person.TeamID > 0 {
		return comm.CodeAlreadyInTeam
	}
	if person.JoinOp == 0 {
		return comm.CodeNoJoinChance
	}

	team, err := teamRepo.FindTeamByID(ctx, h.Request.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}
	if team.Password != h.Request.Password {
		return comm.CodePasswordWrong
	}
	if team.Submit != 0 {
		return comm.CodeTeamSubmitted
	}
	maxTeamSize := 6
	if comm.BizConf != nil && comm.BizConf.MaxTeamSize > 0 {
		maxTeamSize = comm.BizConf.MaxTeamSize
	}
	if int(team.Num) >= maxTeamSize {
		return comm.CodeTeamFull
	}

	err = query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)
		txTeamRepo := repo.NewTeamRepoWithTx(tx)

		updated, errTx := txTeamRepo.IncrementNumIfAvailable(ctx, team.ID, maxTeamSize)
		if errTx != nil {
			return errTx
		}
		if !updated {
			return errTeamJoinConflict
		}
		if errTx := txPeopleRepo.UpdateByOpenID(ctx, openID, map[string]any{
			"team_id": team.ID,
			"role":    comm.RoleMember,
			"join_op": person.JoinOp - 1,
		}); errTx != nil {
			return errTx
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errTeamJoinConflict) {
			latestTeam, latestErr := teamRepo.FindTeamByID(ctx, h.Request.TeamID)
			if latestErr != nil {
				return comm.CodeDatabaseError
			}
			if latestTeam == nil {
				return comm.CodeDataNotFound
			}
			if latestTeam.Submit != 0 {
				return comm.CodeTeamSubmitted
			}
			if int(latestTeam.Num) >= maxTeamSize {
				return comm.CodeTeamFull
			}
			return comm.CodeDataConflict
		}
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func hfTeamJoin(ctx *gin.Context) {
	api := &TeamJoinApi{}
	err := api.Init(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Success(ctx, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
