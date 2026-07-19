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

func TeamDisbandHandler() gin.HandlerFunc {
	api := TeamDisbandApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamDisband).Pointer()).Name()] = api
	return hfTeamDisband
}

type TeamDisbandApi struct {
	Info     struct{} `name:"解散团队" desc:"队长解散团队"`
	Request  struct{}
	Response struct{}
}

func (h *TeamDisbandApi) Init(ctx *gin.Context) error { return nil }
func (h *TeamDisbandApi) Run(ctx *gin.Context) kit.Code {
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
	if err := repo.NewTeamRepo().DisbandTeam(ctx, team.ID); err != nil {
		return comm.CodeServerError
	}
	senderID := person.ID
	messageRepo := repo.NewMessageRepo()
	for _, member := range members {
		if member != nil {
			_ = messageRepo.CreateMessage(ctx, &senderID, member.ID, team.Name+"已经被解散")
		}
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
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
