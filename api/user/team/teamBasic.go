package team

import (
	"reflect"
	"runtime"

	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
)

func TeamBasicHandler() gin.HandlerFunc {
	api := TeamBasicApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamBasic).Pointer()).Name()] = api
	return hfTeamBasic
}

type TeamBasicApi struct {
	Info     struct{} `name:"查询团队基本信息"`
	Request  TeamBasicApiRequest
	Response TeamBasicApiResponse
}

type TeamBasicApiRequest struct {
	Query struct {
		TeamID int64 `form:"team_id" desc:"队伍ID" binding:"required"`
	}
}

type TeamBasicApiResponse struct {
	Name      string `json:"name" desc:"队伍名称"`
	RouteName string `json:"route_name" desc:"队伍路线"`
	Num       uint8  `json:"num" desc:"队伍人数"`
	Slogan    string `json:"slogan" desc:"队伍口号"`
}

func (h *TeamBasicApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindQuery(&h.Request.Query)
}

func (h *TeamBasicApi) Run(ctx *gin.Context) kit.Code {
	team, err := repo.NewTeamRepo().FindTeamByID(ctx, h.Request.Query.TeamID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询团队基本信息失败")
		return comm.CodeServerError
	}
	if team == nil {
		return comm.CodeTeamNotFound
	}
	h.Response = TeamBasicApiResponse{
		Name:      team.Name,
		RouteName: team.RouteName,
		Num:       team.Num,
		Slogan:    team.Slogan,
	}
	return comm.CodeOK
}

func hfTeamBasic(ctx *gin.Context) {
	api := &TeamBasicApi{}
	if err := api.Init(ctx); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	if code := api.Run(ctx); code == comm.CodeOK {
		reply.Reply(ctx, comm.CodeOK, api.Response)
	} else {
		reply.Fail(ctx, code)
	}
}
