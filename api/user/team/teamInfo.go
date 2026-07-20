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

func TeamInfoHandler() gin.HandlerFunc {
	api := TeamInfoApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamInfo).Pointer()).Name()] = api
	return hfTeamInfo
}

type TeamInfoApi struct {
	Info     struct{} `name:"获取团队信息"`
	Request  struct{}
	Response TeamInfoApiResponse
}

type TeamInfoApiResponse struct {
	ID         int64            `json:"id" desc:"团队ID"`
	Name       string           `json:"name" desc:"团队名称"`
	RouteName  string           `json:"route_name" desc:"路线名称"`
	Password   string           `json:"password" desc:"团队密码"`
	Submitted  bool             `json:"submitted" desc:"是否已提交"`
	AllowMatch bool             `json:"allow_match" desc:"是否允许匹配"`
	Slogan     string           `json:"slogan" desc:"团队口号"`
	PointName  string           `json:"point_name" desc:"点位"`
	Status     string           `json:"status" desc:"团队状态"`
	Leader     *TeamMemberView  `json:"leader" desc:"团队队长"`
	Member     []TeamMemberView `json:"member" desc:"团队成员"`
}

type TeamMemberView struct {
	Name       string      `json:"name" desc:"成员姓名"`
	Gender     int8        `json:"gender" desc:"成员性别"`
	Campus     string      `json:"campus" desc:"校区"`
	ID         int         `json:"id" desc:"成员ID"`
	Type       string      `json:"type" desc:"成员类型"`
	Contact    TeamContact `json:"contact" desc:"联系方式"`
	WalkStatus string      `json:"walk_status" desc:"人员状态"`
}

type TeamContact struct {
	QQ     string `json:"qq" desc:"QQ号码"`
	Wechat string `json:"wechat" desc:"微信号码"`
	Tel    string `json:"tel" desc:"电话号码"`
}

func (h *TeamInfoApi) Init(ctx *gin.Context) error { return nil }
func (h *TeamInfoApi) Run(ctx *gin.Context) kit.Code {
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
	submitted, err := teamSubmitted(ctx, team.ID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询团队提交状态失败")
		return comm.CodeServerError
	}

	h.Response.ID = team.ID
	h.Response.Name = team.Name
	h.Response.RouteName = team.RouteName
	h.Response.Password = team.Password
	h.Response.Submitted = submitted
	h.Response.AllowMatch = team.AllowMatch
	h.Response.Slogan = team.Slogan
	h.Response.PointName = team.LatestPointName
	h.Response.Status = team.Status
	h.Response.Member = make([]TeamMemberView, 0, len(members))
	for _, member := range members {
		if member == nil {
			continue
		}
		view := teamMemberView(member)
		if member.OpenID == team.Captain || member.Role == comm.RoleCaptain {
			h.Response.Leader = &view
			continue
		}
		h.Response.Member = append(h.Response.Member, view)
	}
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
