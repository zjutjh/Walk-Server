package api

import (
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

func TeamDisbandHandler() gin.HandlerFunc {
	api := TeamDisbandApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamDisband).Pointer()).Name()] = api
	return hfTeamDisband
}

type TeamDisbandApi struct {
	Info     struct{} `name:"解散队伍" desc:"队长解散队伍"`
	Request  struct{}
	Response struct{}
}

func (h *TeamDisbandApi) Init(ctx *gin.Context) (err error) {
	return err
}

func (h *TeamDisbandApi) Run(ctx *gin.Context) kit.Code {
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
	if person == nil || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}
	if person.Role != comm.RoleCaptain {
		return comm.CodeNotCaptain
	}

	team, err := teamRepo.FindTeamByID(ctx, person.TeamID)
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

		if errTx := txPeopleRepo.UpdateByTeamID(ctx, team.ID, map[string]any{
			"team_id": -1,
			"role":    comm.RoleUnbind,
		}); errTx != nil {
			return errTx
		}
		if errTx := txTeamRepo.DeleteByID(ctx, team.ID); errTx != nil {
			return errTx
		}
		return nil
	})
	if err != nil {
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func hfTeamDisband(ctx *gin.Context) {
	api := &TeamDisbandApi{}
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
