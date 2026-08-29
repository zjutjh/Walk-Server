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

func TeamUpdateHandler() gin.HandlerFunc {
	api := TeamUpdateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamUpdate).Pointer()).Name()] = api
	return hfTeamUpdate
}

type TeamUpdateApi struct {
	Info     struct{} `name:"修改团队"`
	Request  TeamUpdateApiRequest
	Response struct{}
}

type TeamUpdateApiRequest struct {
	Body struct {
		Name       string `json:"name" desc:"队伍名称" binding:"required"`
		RouteName  string `json:"route_name" desc:"团队所属路线" binding:"required"`
		Password   string `json:"password" desc:"团队加入密码" binding:"required"`
		Slogan     string `json:"slogan" desc:"团队标语" binding:"required"`
		AllowMatch *bool  `json:"allow_match" desc:"是否允许随机匹配" binding:"required"`
	}
}

func (h *TeamUpdateApi) Init(ctx *gin.Context) error { return ctx.ShouldBindJSON(&h.Request.Body) }
func (h *TeamUpdateApi) Run(ctx *gin.Context) kit.Code {
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
	if team.Submit && !comm.IsInBizPhase(comm.PhaseAdjustment) {
		return comm.CodeTeamSubmitted
	}
	teamRepo := repo.NewTeamRepo()
	route, err := teamRepo.FindRouteByName(ctx, h.Request.Body.RouteName)
	if err != nil {
		return comm.CodeServerError
	}
	if route == nil {
		return comm.CodeParameterInvalid
	}
	existing, err := teamRepo.FindByNameExceptID(ctx, h.Request.Body.Name, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if existing != nil {
		return comm.CodeTeamNameDuplicated
	}
	passwordChanged := team.Password != h.Request.Body.Password
	routeChanged := team.RouteName != h.Request.Body.RouteName
	memberIDs := make([]int64, 0)
	if passwordChanged || routeChanged {
		members, err := repo.NewPeopleRepo().FindPeopleByTeamID(ctx, team.ID)
		if err != nil {
			return comm.CodeServerError
		}
		for _, member := range members {
			if member != nil && member.ID != person.ID {
				memberIDs = append(memberIDs, member.ID)
			}
		}
	}
	if err := teamRepo.UpdateByID(ctx, team.ID, map[string]any{
		"name":        h.Request.Body.Name,
		"route_name":  h.Request.Body.RouteName,
		"password":    h.Request.Body.Password,
		"slogan":      h.Request.Body.Slogan,
		"allow_match": *h.Request.Body.AllowMatch,
	}); err != nil {
		return comm.CodeServerError
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	if err := teamCache.SetTeamChangeNotice(ctx, memberIDs, passwordChanged, routeChanged); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("记录团队变更通知失败")
	}
	return comm.CodeOK
}

func hfTeamUpdate(ctx *gin.Context) {
	api := &TeamUpdateApi{}
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
