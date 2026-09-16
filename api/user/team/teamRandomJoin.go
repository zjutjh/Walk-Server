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
	if code := comm.CheckBizPhase(comm.PhaseRegistration, comm.PhaseSubmission, comm.PhaseAdjustment); code != comm.CodeOK {
		return code
	}
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	if !(person == nil || person.Role == comm.RoleUnbind || person.TeamID <= 0) {
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
	if !team.AllowMatch {
		return comm.CodeTeamNotAllowMatch
	}
	if int(team.Num) >= comm.BizConf.MaxTeamSize || !team.AllowMatch {
		return comm.CodeTeamFull
	}

	joined, err := teamRepo.JoinTeam(ctx, team.ID, person, true, comm.BizConf.MaxTeamSize)
	if err != nil {
		return comm.CodeServerError
	}
	if !joined {
		return comm.CodeJoinTeamFailed
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	return comm.CodeOK
}

func hfTeamRandomJoin(ctx *gin.Context) {
	api := &TeamRandomJoinApi{}
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
