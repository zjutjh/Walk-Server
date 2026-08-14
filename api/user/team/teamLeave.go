package team

import (
	"reflect"
	"runtime"

	peopleCache "app/dao/cache/people"
	teamCache "app/dao/cache/team"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func TeamLeaveHandler() gin.HandlerFunc {
	api := TeamLeaveApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamLeave).Pointer()).Name()] = api
	return hfTeamLeave
}

type TeamLeaveApi struct {
	Info     struct{} `name:"离开团队"`
	Request  struct{}
	Response struct{}
}

func (h *TeamLeaveApi) Init(ctx *gin.Context) error { return nil }
func (h *TeamLeaveApi) Run(ctx *gin.Context) kit.Code {
	if code := comm.CheckBizPhase(comm.PhaseRegistration, comm.PhaseAdjustment); code != comm.CodeOK {
		return code
	}
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	if person.Role == comm.RoleCaptain || team.Captain == person.ID {
		return comm.CodeCannotLeaveTeam
	}
	submitted, err := teamCache.IsTeamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if submitted && !comm.IsInBizPhase(comm.PhaseAdjustment) {
		return comm.CodeTeamSubmitted
	}
	ok, err := repo.NewTeamRepo().RemoveMember(ctx, team.ID, person)
	if err != nil {
		return comm.CodeServerError
	}
	if !ok {
		return comm.CodeLeaveTeamFailed
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	_ = peopleCache.DelPersonByID(ctx, person.ID)
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
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
