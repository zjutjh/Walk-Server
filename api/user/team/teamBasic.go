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
	Info     struct{} `name:"团队邀请预览"`
	Request  TeamBasicApiRequest
	Response TeamBasicApiResponse
}

type TeamBasicApiRequest struct {
	Body struct {
		TeamID   int64  `json:"team_id" desc:"队伍ID" binding:"required"`
		Password string `json:"password" desc:"队伍密码" binding:"required"`
	}
}

type TeamBasicApiResponse struct {
	Name           string         `json:"name" desc:"队伍名称"`
	RouteName      string         `json:"route_name" desc:"队伍路线"`
	Slogan         string         `json:"slogan" desc:"队伍口号"`
	CaptainName    string         `json:"captain_name" desc:"脱敏后的队长姓名"`
	IsFull         bool           `json:"is_full" desc:"队伍是否已满"`
	MemberCount    *uint8         `json:"member_count" desc:"当前人数，未登录时为null"`
	MaxMemberCount *int           `json:"max_member_count" desc:"人数上限，未登录时为null"`
	Phase          *comm.BizPhase `json:"phase" desc:"当前阶段，未登录时为null"`
}

func (h *TeamBasicApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request.Body)
}

func (h *TeamBasicApi) Run(ctx *gin.Context) kit.Code {
	team, err := repo.NewTeamRepo().FindTeamByID(ctx, h.Request.Body.TeamID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询团队基本信息失败")
		return comm.CodeServerError
	}
	if team == nil {
		return comm.CodeTeamNotFound
	}
	if team.Password != h.Request.Body.Password {
		return comm.CodePasswordWrong
	}

	captain, err := repo.NewPeopleRepo().FindPeopleByID(ctx, team.Captain)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询队长信息失败")
		return comm.CodeServerError
	}
	if captain == nil {
		nlog.Pick().WithContext(ctx).Warn("团队队长不存在")
		return comm.CodeServerError
	}

	h.Response = TeamBasicApiResponse{
		Name:        team.Name,
		RouteName:   team.RouteName,
		Slogan:      team.Slogan,
		CaptainName: comm.MaskPersonName(captain.Name),
		IsFull:      int(team.Num) >= comm.BizConf.MaxTeamSize,
	}

	if _, err := comm.GetUserIDFromCtx(ctx); err == nil {
		memberCount := team.Num
		maxMemberCount := comm.BizConf.MaxTeamSize
		h.Response.MemberCount = &memberCount
		h.Response.MaxMemberCount = &maxMemberCount
		if phase := comm.CurrentBizPhase(); phase != "" {
			h.Response.Phase = &phase
		}
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
