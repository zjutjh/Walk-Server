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

func TeamLeaveHandler() gin.HandlerFunc {
	api := TeamLeaveApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamLeave).Pointer()).Name()] = api
	return hfTeamLeave
}

type TeamLeaveApi struct {
	Info     struct{} `name:"退出队伍" desc:"普通队员退出队伍"`
	Request  struct{}
	Response struct{}
}

func (h *TeamLeaveApi) Init(ctx *gin.Context) (err error) {
	return err
}

func (h *TeamLeaveApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}
	if person.Role == comm.RoleCaptain {
		return comm.CodePermissionDenied
	}

	team, err := teamRepo.FindByID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}
	if team.Submit != 0 {
		return comm.CodeTeamSubmitted
	}

	err = query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)
		txTeamRepo := repo.NewTeamRepoWithTx(tx)

		updated, errTx := txTeamRepo.DecrementNumIfPositive(ctx, team.ID)
		if errTx != nil {
			return errTx
		}
		if !updated {
			return errTeamLeaveConflict
		}
		if errTx := txPeopleRepo.UpdateByOpenID(ctx, openID, map[string]any{
			"team_id": -1,
			"role":    comm.RoleUnbind,
		}); errTx != nil {
			return errTx
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errTeamLeaveConflict) {
			return comm.CodeDataConflict
		}
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func hfTeamLeave(ctx *gin.Context) {
	api := &TeamLeaveApi{}
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
