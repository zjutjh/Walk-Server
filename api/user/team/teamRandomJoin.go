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

func TeamRandomJoinHandler() gin.HandlerFunc {
	api := TeamRandomJoinApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamRandomJoin).Pointer()).Name()] = api
	return hfTeamRandomJoin
}

type TeamRandomJoinApi struct {
	Info     struct{} `name:"随机加入团队" `
	Request  TeamRandomJoinApiRequest
	Response struct{}
}

type TeamRandomJoinApiRequest struct {
	Body struct {
		ID int64 `json:"id" desc:"队伍ID" binding:"required"`
	}
}

func (h *TeamRandomJoinApi) Init(ctx *gin.Context) error { return ctx.ShouldBindJSON(&h.Request.Body) }
func (h *TeamRandomJoinApi) Run(ctx *gin.Context) kit.Code {
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
	team, err := teamRepo.FindTeamByID(ctx, h.Request.Body.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if team == nil {
		return comm.CodeTeamNotFound
	}
	submitted, err := teamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if submitted {
		return comm.CodeTeamSubmitted
	}
	if !team.AllowMatch {
		return comm.CodeTeamNotAllowMatch
	}
	if int(team.Num) >= comm.BizConf.MaxTeamSize || !team.AllowMatch {
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
		return comm.CodeTeacherCannotJoinStudentTeam
	}

	joined, err := teamRepo.JoinTeam(ctx, team.ID, person, true, comm.BizConf.MaxTeamSize)
	if err != nil {
		return comm.CodeServerError
	}
	if !joined {
		return comm.CodeJoinTeamFailed
	}
	senderID := person.ID
	messageRepo := repo.NewMessageRepo()
	for _, member := range members {
		if member == nil {
			continue
		}
		_ = messageRepo.CreateMessage(ctx, &senderID, member.ID, person.Name+"通过随机组队加入了队伍")
	}
	return comm.CodeOK
}

func hfTeamRandomJoin(ctx *gin.Context) {
	api := &TeamRandomJoinApi{}
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
