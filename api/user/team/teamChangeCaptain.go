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

func TeamChangeCaptainHandler() gin.HandlerFunc {
	api := TeamChangeCaptainApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamChangeCaptain).Pointer()).Name()] = api
	return hfTeamChangeCaptain
}

type TeamChangeCaptainApi struct {
	Info     struct{} `name:"更换队长" desc:"队长转让队长身份"`
	Request  TeamChangeCaptainApiRequest
	Response struct{}
}

type TeamChangeCaptainApiRequest struct {
	Body struct {
		ID int64 `json:"id" desc:"更换的队长ID" binding:"required"`
	}
}

func (h *TeamChangeCaptainApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request.Body)
}
func (h *TeamChangeCaptainApi) Run(ctx *gin.Context) kit.Code {
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
	newCaptain, err := repo.NewPeopleRepo().FindPeopleByID(ctx, h.Request.Body.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if newCaptain == nil {
		return comm.CodePeopleNotFound
	}
	if newCaptain.TeamID != team.ID {
		return comm.CodePermissionDenied
	}
	if person.Type != comm.MemberTypeStudent && newCaptain.Type == comm.MemberTypeStudent {
		return comm.CodePermissionDenied
	}
	if err := repo.NewTeamRepo().ChangeCaptain(ctx, team.ID, person.OpenID, newCaptain.OpenID); err != nil {
		return comm.CodeServerError
	}
	return comm.CodeOK
}

func hfTeamChangeCaptain(ctx *gin.Context) {
	api := &TeamChangeCaptainApi{}
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
