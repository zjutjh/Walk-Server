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
	Body struct {
		ID int64 `json:"id" desc:"成员ID" binding:"required"`
	}
}

func (h *TeamRemoveMemberApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request.Body)
}
func (h *TeamRemoveMemberApi) Run(ctx *gin.Context) kit.Code {
	if code := comm.CheckBizPhase(comm.PhaseRegistration, comm.PhaseSubmission, comm.PhaseAdjustment); code != comm.CodeOK {
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
	if !(person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.ID) {
		return comm.CodeNotCaptain
	}
	removed, err := repo.NewPeopleRepo().FindPeopleByID(ctx, h.Request.Body.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if removed == nil || removed.TeamID != team.ID {
		return comm.CodePeopleNotFound
	}
	if removed.ID == person.ID {
		return comm.CodeCannotRemoveSelf
	}

	ok, err := repo.NewTeamRepo().RemoveMember(ctx, team.ID, removed)
	if err != nil {
		return comm.CodeServerError
	}
	if !ok {
		return comm.CodeRemoveFailed
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	return comm.CodeOK
}

func hfTeamRemoveMember(ctx *gin.Context) {
	api := &TeamRemoveMemberApi{}
	if err := api.Init(ctx); err != nil {
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
