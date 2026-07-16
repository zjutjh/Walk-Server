package team

import (
	"reflect"
	"runtime"

	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func TeamJoinHandler() gin.HandlerFunc {
	api := TeamJoinApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamJoin).Pointer()).Name()] = api
	return hfTeamJoin
}

type TeamJoinApi struct {
	Info     struct{} `name:"加入团队" `
	Request  TeamJoinApiRequest
	Response struct{}
}

type TeamJoinApiRequest struct {
	Body struct {
		TeamID       int64  `json:"team_id" desc:"队伍ID" binding:"required"`
		Password string `json:"password" desc:"团队加入密码" binding:"required"`
	}
}

func (h *TeamJoinApi) Init(ctx *gin.Context) error   { return ctx.ShouldBindJSON(&h.Request.Body) }
func (h *TeamJoinApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	if !isUnbound(person) {
		return comm.CodeAlreadyInTeam
	}
	if person.JoinOp == 0 {
		return comm.CodeNoJoinChance
	}

	teamRepo := repo.NewTeamRepo()
	team, err := teamRepo.FindTeamByID(ctx, h.Request.Body.TeamID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询队伍失败")
		return comm.CodeServerError
	}
	if team == nil {
		return comm.CodeTeamNotFound
	}
	if team.Password != h.Request.Body.Password {
		return comm.CodePasswordWrong
	}
	submitted, err := teamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if submitted {
		return comm.CodeTeamSubmitted
	}
	if int(team.Num) >= comm.BizConf.MaxTeamSize {
		return comm.CodeTeamFull
	}

	members, err := repo.NewPeopleRepo().FindPeopleByTeamID(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	var captain *model.People
	for _, member := range members {
		if member != nil && member.OpenID == team.Captain {
			captain = member
			break
		}
	}
	if !canTeacherJoinTeam(captain, person) {
		return comm.CodePermissionDenied
	}

	joined, err := teamRepo.JoinTeam(ctx, team.ID, person, true, comm.BizConf.MaxTeamSize)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("加入团队失败")
		return comm.CodeServerError
	}
	if !joined {
		return comm.CodeTeamFull
	}
	for _, member := range members {
		sendTeamMessage(ctx, person, member, person.Name+"加入了团队")
	}
	return comm.CodeOK
}

func hfTeamJoin(ctx *gin.Context) {
	api := &TeamJoinApi{}
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
