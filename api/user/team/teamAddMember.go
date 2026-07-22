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

func TeamAddMemberHandler() gin.HandlerFunc {
	api := TeamAddMemberApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamAddMember).Pointer()).Name()] = api
	return hfTeamAddMember
}

type TeamAddMemberApi struct {
	Info     struct{} `name:"添加成员" desc:"队长按学号添加成员"`
	Request  TeamAddMemberApiRequest
	Response struct{}
}

type TeamAddMemberApiRequest struct {
	Query struct {
		StuID string `form:"stuid" desc:"学号" binding:"required"`
	}
}

func (h *TeamAddMemberApi) Init(ctx *gin.Context) error { return ctx.ShouldBindQuery(&h.Request.Query) }
func (h *TeamAddMemberApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	if !(person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.OpenID) {
		return comm.CodeNotCaptain
	}
	submitted, err := teamCache.IsTeamSubmitted(ctx, team.ID)
	if err != nil {
		return comm.CodeServerError
	}
	if submitted {
		return comm.CodeTeamSubmitted
	}
	if int(team.Num) >= comm.BizConf.MaxTeamSize {
		return comm.CodeTeamFull
	}
	if h.Request.Query.StuID == person.StuID {
		return comm.CodeCannotAddSelf
	}

	peopleRepo := repo.NewPeopleRepo()
	newMember, err := peopleRepo.FindPeopleByStuID(ctx, h.Request.Query.StuID)
	if err != nil {
		return comm.CodeServerError
	}
	if newMember == nil {
		return comm.CodePeopleNotFound
	}
	if !(newMember == nil || newMember.Role == comm.RoleUnbind || newMember.TeamID <= 0) {
		return comm.CodeAlreadyInTeam
	}
	if person != nil && newMember != nil && person.Type == comm.MemberTypeStudent && newMember.Type == comm.MemberTypeTeacher {
		return comm.CodeTeacherCannotJoinStudentTeam
	}

	joined, err := repo.NewTeamRepo().JoinTeam(ctx, team.ID, newMember, false, comm.BizConf.MaxTeamSize)
	if err != nil {
		return comm.CodeServerError
	}
	if !joined {
		return comm.CodeTeamFull
	}
	senderID := person.ID
	messageRepo := repo.NewMessageRepo()
	_ = messageRepo.CreateMessage(ctx, &senderID, newMember.ID, "你被"+person.Name+"添加至团队"+team.Name)
	_ = messageRepo.CreateMessage(ctx, nil, person.ID, "你添加了成员"+newMember.Name)
	return comm.CodeOK
}

func hfTeamAddMember(ctx *gin.Context) {
	api := &TeamAddMemberApi{}
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
