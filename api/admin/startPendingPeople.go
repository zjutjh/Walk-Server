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
	teamCache "app/dao/cache/team"
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
	admin, ok := repo.GetAdminInfo(ctx)
	if !ok {
		return comm.CodeNotLoggedIn
	}

	var teamIDs []int64
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		peopleRepo := repo.NewPeopleRepoWithTx(tx)
		teamRepo := repo.NewTeamRepoWithTx(tx)

		var err error
		teamIDs, err = peopleRepo.UpdatePendingWalkStatusByCampus(ctx, admin.Campus)
		if err != nil {
			return err
		}
		return teamRepo.UpdateStatusByIDs(ctx, teamIDs, comm.TeamStatusInProgress)
	})
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("批量更新待出发人员状态失败")
		return comm.CodeServerError
	}
	for _, teamID := range teamIDs {
		_ = teamCache.DelTeamByID(ctx, teamID)
		_ = teamCache.DeleteTeamInfo(ctx, teamID)
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
