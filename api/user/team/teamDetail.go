package team

import (
	"reflect"
	"runtime"

	"app/comm"
	teamCache "app/dao/cache/team"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
)

func TeamDetailHandler() gin.HandlerFunc {
	api := TeamDetailApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamDetail).Pointer()).Name()] = api
	return hfTeamDetail
}

type TeamDetailApi struct {
	Info     struct{} `name:"团队详细信息"`
	Request  struct{}
	Response TeamDetailResponse
}

type TeamDetailResponse struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Slogan          string `json:"slogan"`
	Code            string `json:"code"`
	Password        string `json:"password" desc:"团队加入密码，普通队员返回空字符串"`
	RouteName       string `json:"route_name"`
	Submitted       bool   `json:"submitted"`
	AllowMatch      bool   `json:"allow_match"`
	Status          string `json:"status"`
	LatestPointName string `json:"latest_point_name"`
}

func (h *TeamDetailApi) Init(ctx *gin.Context) error { return nil }

func (h *TeamDetailApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	submitted, err := teamCache.IsTeamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	password := ""
	if person.Role == comm.RoleCaptain && team.Captain == person.ID {
		password = team.Password
	}
	h.Response = TeamDetailResponse{
		ID: team.ID, Name: team.Name, Slogan: team.Slogan, Code: team.Code, Password: password,
		RouteName: team.RouteName, Submitted: submitted || team.Submit, AllowMatch: team.AllowMatch,
		Status: team.Status, LatestPointName: team.LatestPointName,
	}
	return comm.CodeOK
}

func hfTeamDetail(ctx *gin.Context) {
	api := &TeamDetailApi{}
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
