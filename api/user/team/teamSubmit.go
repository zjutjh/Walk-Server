package team

import (
	"reflect"
	"runtime"

	teamCache "app/dao/cache/team"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func TeamSubmitHandler() gin.HandlerFunc {
	api := TeamSubmitApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamSubmit).Pointer()).Name()] = api
	return hfTeamSubmit
}

type TeamSubmitApi struct {
	Info     struct{} `name:"提交团队" desc:"提交团队"`
	Request  struct{}
	Response struct{}
}

func (h *TeamSubmitApi) Init(ctx *gin.Context) error { return nil }
func (h *TeamSubmitApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	if !isCaptain(person, team) {
		return comm.CodeNotCaptain
	}
	if team.Num < 4 {
		return comm.CodeTeamNotEnough
	}
	day, routeCode, code := teamQuotaRoute(ctx, team.RouteName)
	if code != comm.CodeOK {
		return code
	}
	result, err := teamCache.SubmitTeam(ctx, team.ID, day, routeCode)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("提交团队扣减名额失败")
		return comm.CodeServerError
	}
	switch result {
	case 1:
		return comm.CodeTeamSubmitted
	case 2:
		return comm.CodeUserNoQuota
	}
	if err := repo.NewTeamRepo().UpdateByID(ctx, team.ID, map[string]any{"submit": true}); err != nil {
		_, _ = teamCache.RollbackTeamSubmit(ctx, team.ID, day, routeCode)
		return comm.CodeServerError
	}
	return comm.CodeOK
}

func hfTeamSubmit(ctx *gin.Context) {
	api := &TeamSubmitApi{}
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
