package team

import (
	"reflect"
	"runtime"

	"app/comm"
	teamCache "app/dao/cache/team"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
)

func TeamMemberHandler() gin.HandlerFunc {
	api := TeamMemberApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamMember).Pointer()).Name()] = api
	return hfTeamMember
}

type TeamMemberApi struct {
	Info     struct{} `name:"队员详细信息" desc:"查询当前团队中的指定成员"`
	Request  TeamMemberRequest
	Response TeamMemberResponse
}

type TeamMemberRequest struct {
	Query struct {
		ID int64 `form:"id" desc:"成员ID" binding:"required"`
	}
}

type TeamMemberResponse struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Type               string `json:"type"`
	Role               string `json:"role"`
	Tel                string `json:"tel"`
	Wechat             string `json:"wechat"`
	QQ                 string `json:"qq"`
	WalkStatus         string `json:"walk_status"`
	CanRemove          bool   `json:"can_remove"`
	CanTransferCaptain bool   `json:"can_transfer_captain"`
}

func (h *TeamMemberApi) Init(ctx *gin.Context) error { return ctx.ShouldBindQuery(&h.Request.Query) }

func (h *TeamMemberApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	member, err := repo.NewPeopleRepo().FindPeopleByID(ctx, h.Request.Query.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if member == nil || member.TeamID != team.ID {
		return comm.CodePeopleNotFound
	}
	submitted, err := teamCache.IsTeamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	isCaptain := person.Role == comm.RoleCaptain && team.Captain == person.ID
	isSelf := person.ID == member.ID
	canOperate := isCaptain && !isSelf && !submitted && !team.Submit
	role := member.Role
	if member.ID == team.Captain {
		role = comm.RoleCaptain
	}
	h.Response = TeamMemberResponse{
		ID: member.ID, Name: member.Name, Type: member.Type, Role: role,
		Tel: member.Tel, Wechat: member.Wechat, QQ: member.Qq, WalkStatus: member.WalkStatus,
		CanRemove: canOperate, CanTransferCaptain: canOperate,
	}
	return comm.CodeOK
}

func hfTeamMember(ctx *gin.Context) {
	api := &TeamMemberApi{}
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
