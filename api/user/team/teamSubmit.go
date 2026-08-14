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
	if !(person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.ID) {
		return comm.CodeNotCaptain
	}
	if team.Num < 4 {
		return comm.CodeTeamNotEnough
	}
	day, code := teamQuotaDay(ctx)
	if code != comm.CodeOK {
		return code
	}
	result, err := teamCache.SubmitTeam(ctx, team.ID, day)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("提交团队扣减名额失败")
		return comm.CodeServerError
	}
	switch result {
	case 1:
		return comm.CodeTeamSubmitted
	case 2:
		return comm.CodeDailyQuotaFull
	case 3:
		return comm.CodeActivityQuotaFull
	}
	if err := repo.NewTeamRepo().UpdateByID(ctx, team.ID, map[string]any{"submit": true}); err != nil {
		_, _ = teamCache.RollbackTeamSubmit(ctx, team.ID, day)
		return comm.CodeServerError
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	return comm.CodeOK
}

func teamQuotaDay(ctx *gin.Context) (int, kit.Code) {
	if !comm.IsInRegisterTime() {
		return 0, comm.CodeNotInRegisterTime
	}
	day := comm.CurrentActivityDay()
	if _, ok := comm.DailyTeamLimit(day); !ok {
		nlog.Pick().WithContext(ctx).Warn("未配置当天提交名额")
		return 0, comm.CodeNotInRegisterTime
	}
	return day, comm.CodeOK
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
