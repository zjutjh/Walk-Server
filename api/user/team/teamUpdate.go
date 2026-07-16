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

func (h *TeamUpdateApi) Init(ctx *gin.Context) error   { return ctx.ShouldBindJSON(&h.Request.Body) }
func (h *TeamUpdateApi) Run(ctx *gin.Context) kit.Code {
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
	if err := teamRepo.UpdateByID(ctx, team.ID, map[string]any{
		"name":        h.Request.Body.Name,
		"route_name":  h.Request.Body.RouteName,
		"password":    h.Request.Body.Password,
		"slogan":      h.Request.Body.Slogan,
		"allow_match": *h.Request.Body.AllowMatch,
	}); err != nil {
		return comm.CodeServerError
	}
	return comm.CodeOK
}

func hfTeamUpdate(ctx *gin.Context) {
	api := &TeamUpdateApi{}
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
