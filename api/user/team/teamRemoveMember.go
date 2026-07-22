package team

import (
	"reflect"
	"runtime"

	"app/dao/repo"

	teamCache "app/dao/cache/team"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func TeamRemoveMemberHandler() gin.HandlerFunc {
	api := TeamRemoveMemberApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamRemoveMember).Pointer()).Name()] = api
	return hfTeamRemoveMember
}

type TeamRemoveMemberApi struct {
	Info     struct{} `name:"移除成员" `
	Request  TeamRemoveMemberApiRequest
	Response struct{}
}

type TeamRemoveMemberApiRequest struct {
	Query struct {
		ID int `form:"id" desc:"成员ID" binding:"required"`
	}
}

func (h *TeamRemoveMemberApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindQuery(&h.Request.Query)
}
func (h *TeamRemoveMemberApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	if !(person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.OpenID) {
		return comm.CodeNotCaptain
	}
	submitted, err := teamCache.IsTeamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if submitted {
		return comm.CodeTeamSubmitted
	}
	removed, err := repo.NewPeopleRepo().FindPeopleByID(ctx, int64(h.Request.Query.ID))
	if err != nil {
		return comm.CodeServerError
	}
	if removed == nil {
		return comm.CodePeopleNotFound
	}
	if removed.OpenID == person.OpenID {
		return comm.CodeCannotRemoveSelf
	}

	ok, err := repo.NewTeamRepo().RemoveMember(ctx, team.ID, removed)
	if err != nil {
		return comm.CodeServerError
	}
	if !ok {
		return comm.CodeRemoveFailed
	}
	messageRepo := repo.NewMessageRepo()
	_ = messageRepo.CreateMessage(ctx, nil, removed.ID, "你被团队"+team.Name+"踢出")
	_ = messageRepo.CreateMessage(ctx, nil, person.ID, "你踢出了成员"+removed.Name)
	return comm.CodeOK
}

func hfTeamRemoveMember(ctx *gin.Context) {
	api := &TeamRemoveMemberApi{}
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
