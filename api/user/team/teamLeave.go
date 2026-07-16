package team

import (
	"reflect"
	"runtime"

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

func (h *TeamLeaveApi) Init(ctx *gin.Context) error   { return nil }
func (h *TeamLeaveApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	if person.Role == comm.RoleCaptain || team.Captain == person.OpenID {
		return comm.CodePermissionDenied
	}
	submitted, err := teamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if submitted {
		return comm.CodeTeamSubmitted
	}
	members, err := repo.NewPeopleRepo().FindPeopleByTeamID(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	ok, err := repo.NewTeamRepo().RemoveMember(ctx, team.ID, person)
	if err != nil {
		return comm.CodeServerError
	}
	if !ok {
		return comm.CodeDataConflict
	}
	for _, member := range members {
		if member != nil && member.OpenID != person.OpenID {
			sendTeamMessage(ctx, person, member, person.Name+"已经离开了队伍")
		}
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
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
