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

func TeamRandomListHandler() gin.HandlerFunc {
	api := TeamRandomListApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamRandomList).Pointer()).Name()] = api
	return hfTeamRandomList
}

type TeamRandomListApi struct {
	Info     struct{} `name:"随机组队列表" desc:"随机获取开放匹配团队"`
	Request  TeamRandomListApiRequest
	Response TeamRandomListApiResponse
}

type TeamRandomListApiRequest struct {
	Query struct {
		RouteName string `form:"route_name" desc:"团队所属路线" binding:"required"`
	}
}

type TeamRandomListApiResponse struct {
	Teams []TeamRandomListItem `json:"teams" desc:"随机匹配团队列表"`
}

type TeamRandomListItem struct {
	ID        int64  `json:"id" desc:"队伍ID"`
	Name      string `json:"name" desc:"队伍名称"`
	Num       uint8  `json:"num" desc:"团队人数"`
	Slogan    string `json:"slogan" desc:"团队标语"`
	RouteName string `json:"route_name" desc:"团队所属路线"`
}

func (h *TeamRandomListApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindQuery(&h.Request.Query)
}
func (h *TeamRandomListApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()
	route, err := teamRepo.FindRouteByName(ctx, h.Request.Query.RouteName)
	if err != nil {
		return comm.CodeServerError
	}
	if route == nil {
		return comm.CodeParameterInvalid
	}
	teams, err := teamRepo.ListRandomMatchTeams(ctx, h.Request.Query.RouteName, comm.BizConf.MaxTeamSize)
	if err != nil {
		return comm.CodeServerError
	}
	if len(teams) == 0 {
		return comm.CodeNotFindTeams
	}
	h.Response.Teams = make([]TeamRandomListItem, 0, len(teams))
	for _, team := range teams {
		h.Response.Teams = append(h.Response.Teams, TeamRandomListItem{
			ID:        team.ID,
			Name:      team.Name,
			Num:       team.Num,
			Slogan:    team.Slogan,
			RouteName: team.RouteName,
		})
	}
	return comm.CodeOK
}

func hfTeamRandomList(ctx *gin.Context) {
	api := &TeamRandomListApi{}
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
