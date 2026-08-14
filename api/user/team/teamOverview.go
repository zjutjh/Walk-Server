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

func TeamOverviewHandler() gin.HandlerFunc {
	api := TeamOverviewApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamOverview).Pointer()).Name()] = api
	return hfTeamOverview
}

type TeamOverviewApi struct {
	Info     struct{} `name:"团队页面基本信息" desc:"返回团队摘要和成员列表"`
	Request  struct{}
	Response TeamOverviewResponse
}

type TeamOverviewResponse struct {
	Team    TeamSummary         `json:"team"`
	Members []TeamMemberSummary `json:"members"`
}

type TeamSummary struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slogan    string `json:"slogan"`
	RouteName string `json:"route_name"`
}

type TeamMemberSummary struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Role string `json:"role"`
}

func (h *TeamOverviewApi) Init(ctx *gin.Context) error { return nil }

func (h *TeamOverviewApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	members, err := repo.NewPeopleRepo().FindPeopleByTeamID(ctx, team.ID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询团队成员失败")
		return comm.CodeServerError
	}
	h.Response.Team = TeamSummary{ID: team.ID, Name: team.Name, Slogan: team.Slogan, RouteName: team.RouteName}
	h.Response.Members = make([]TeamMemberSummary, 0, len(members))
	for _, member := range members {
		if member == nil {
			continue
		}
		role := member.Role
		if member.ID == team.Captain {
			role = comm.RoleCaptain
		}
		h.Response.Members = append(h.Response.Members, TeamMemberSummary{
			ID: member.ID, Name: member.Name, Type: member.Type, Role: role,
		})
	}
	return comm.CodeOK
}

func hfTeamOverview(ctx *gin.Context) {
	api := &TeamOverviewApi{}
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
