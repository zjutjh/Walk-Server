package api

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/repo"
)

func TeamInfoHandler() gin.HandlerFunc {
	api := TeamInfoApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamInfo).Pointer()).Name()] = api
	return hfTeamInfo
}

type TeamInfoApi struct {
	Info     struct{} `name:"队伍信息" desc:"获取当前队伍与成员信息"`
	Request  struct{}
	Response TeamInfoApiResponse
}

func (h *TeamInfoApi) Init(ctx *gin.Context) (err error) {
	return err
}

func (h *TeamInfoApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindPeopleByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}

	team, err := teamRepo.FindTeamByID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}

	members, err := peopleRepo.ListByTeamID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}

	h.Response.Team = toTeamInfoTeamView(team)
	h.Response.Members = toTeamInfoMemberViews(members)
	return comm.CodeOK
}

func hfTeamInfo(ctx *gin.Context) {
	api := &TeamInfoApi{}
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
