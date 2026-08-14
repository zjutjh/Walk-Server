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
	if !(person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.ID) {
		return comm.CodeNotCaptain
	}
	submitted, err := teamCache.IsTeamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if submitted && !comm.IsInBizPhase(comm.PhaseAdjustment) {
		return comm.CodeTeamSubmitted
	}
	newCaptain, err := repo.NewPeopleRepo().FindPeopleByID(ctx, h.Request.Body.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if newCaptain == nil || newCaptain.TeamID != team.ID {
		return comm.CodePeopleNotFound
	}
	if err := repo.NewTeamRepo().ChangeCaptain(ctx, team.ID, person.ID, newCaptain.ID); err != nil {
		return comm.CodeServerError
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	_ = peopleCache.DelPersonByID(ctx, person.ID)
	_ = peopleCache.DelPersonByID(ctx, newCaptain.ID)
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
