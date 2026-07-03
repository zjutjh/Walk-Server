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
	repo "app/dao/repo"
)

func StartPendingPeopleHandler() gin.HandlerFunc {
	api := StartPendingPeopleApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(startPendingPeople).Pointer()).Name()] = api
	return startPendingPeople
}

type StartPendingPeopleApi struct {
	Info     struct{} `name:"待出发人员批量改为进行中"`
	Request  struct{}
	Response struct{}
}

func (s *StartPendingPeopleApi) Run(ctx *gin.Context) kit.Code {
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)
		txTeamRepo := repo.NewTeamRepoWithTx(tx)

		_, teamIDs, err := txPeopleRepo.UpdateWalkStatusByCurrent(ctx, comm.WalkStatusPending, comm.WalkStatusInProgress)
		if err != nil {
			return err
		}
		if len(teamIDs) > 0 {
			if err := txTeamRepo.UpdateStatusByIDs(ctx, teamIDs, comm.TeamStatusInProgress); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("批量更新待出发人员状态失败")
		return comm.CodeServerError
	}

	return comm.CodeOK
}

func startPendingPeople(ctx *gin.Context) {
	api := &StartPendingPeopleApi{}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
