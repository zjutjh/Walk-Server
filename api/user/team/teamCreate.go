package team

import (
	"reflect"
	"runtime"
	"time"

	teamCache "app/dao/cache/team"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func TeamCreateHandler() gin.HandlerFunc {
	api := TeamCreateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamCreate).Pointer()).Name()] = api
	return hfTeamCreate
}

type TeamCreateApi struct {
	Info     struct{} `name:"创建团队" `
	Request  TeamCreateApiRequest
	Response TeamCreateApiResponse
}

type TeamCreateApiRequest struct {
	Body struct {
		Name       string `json:"name" desc:"队伍名称" binding:"required"`
		RouteName  string `json:"route_name" desc:"团队所属路线" binding:"required"`
		Password   string `json:"password" desc:"团队加入密码" binding:"required"`
		Slogan     string `json:"slogan" desc:"团队标语" binding:"required"`
		AllowMatch *bool  `json:"allow_match" desc:"是否允许随机匹配" binding:"required"`
	}
}

type TeamCreateApiResponse struct {
	TeamID int64 `json:"team_id" desc:"队伍ID"`
}

func (h *TeamCreateApi) Init(ctx *gin.Context) error { return ctx.ShouldBindJSON(&h.Request.Body) }
func (h *TeamCreateApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	if !isUnbound(person) {
		return comm.CodeAlreadyInTeam
	}
	if person.CreatedOp == 0 {
		return comm.CodeNoCreateChance
	}

	teamRepo := repo.NewTeamRepo()
	route, err := teamRepo.FindRouteByName(ctx, h.Request.Body.RouteName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询路线失败")
		return comm.CodeServerError
	}
	if route == nil {
		return comm.CodeParameterInvalid
	}
	existing, err := teamRepo.FindTeamByName(ctx, h.Request.Body.Name)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询队名失败")
		return comm.CodeServerError
	}
	if existing != nil {
		return comm.CodeTeamNameDuplicated
	}

	team := &model.Team{
		Name:       h.Request.Body.Name,
		Num:        1,
		Password:   h.Request.Body.Password,
		Slogan:     h.Request.Body.Slogan,
		AllowMatch: *h.Request.Body.AllowMatch,
		Captain:    person.OpenID,
		Submit:     false,
		RouteName:  h.Request.Body.RouteName,
		Status:     comm.TeamStatusNotStart,
		Time:       time.Now(),
	}
	if err := teamRepo.CreateWithCaptain(ctx, team, person); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("创建团队失败")
		return comm.CodeServerError
	}
	h.Response.TeamID = team.ID
	return comm.CodeOK
}

func currentTeamUser(ctx *gin.Context) (*model.People, kit.Code) {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return nil, comm.CodeNotLoggedIn
	}
	if comm.BizConf.AESSecret != "" {
		if decrypted, err := comm.AesDecrypt(openID, comm.BizConf.AESSecret); err == nil && decrypted != "" {
			openID = decrypted
		}
	}
	person, err := repo.NewPeopleRepo().FindPeopleByOpenID(ctx, openID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询当前用户失败")
		return nil, comm.CodeServerError
	}
	if person == nil {
		return nil, comm.CodePeopleNotFound
	}
	return person, comm.CodeOK
}

func currentUserTeam(ctx *gin.Context, person *model.People) (*model.Team, kit.Code) {
	if isUnbound(person) {
		return nil, comm.CodeNotInTeam
	}
	team, err := repo.NewTeamRepo().FindTeamByID(ctx, person.TeamID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询当前队伍失败")
		return nil, comm.CodeServerError
	}
	if team == nil {
		return nil, comm.CodeTeamNotFound
	}
	return team, comm.CodeOK
}

func teamSubmitted(ctx *gin.Context, teamID int64) (bool, error) {
	return teamCache.IsTeamSubmitted(ctx, teamID)
}

func teamQuotaRoute(ctx *gin.Context, routeName string) (int, int, kit.Code) {
	day := comm.CurrentActivityDay()
	routeCode, ok := comm.RouteQuotaCode(routeName)
	if !ok {
		nlog.Pick().WithContext(ctx).Warn("路线未配置提交名额编号")
		return 0, 0, comm.CodeNotInRegisterTime
	}
	if _, ok := comm.TeamUpperLimit(day, routeCode); !ok {
		nlog.Pick().WithContext(ctx).Warn("未配置当天路线提交名额")
		return 0, 0, comm.CodeNotInRegisterTime
	}
	return day, routeCode, comm.CodeOK
}

func isUnbound(person *model.People) bool {
	return person == nil || person.Role == comm.RoleUnbind || person.TeamID <= 0
}

func isCaptain(person *model.People, team *model.Team) bool {
	return person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.OpenID
}

func canTeacherJoinTeam(captain, member *model.People) bool {
	return !(captain != nil && member != nil && captain.Type == comm.MemberTypeStudent && member.Type == comm.MemberTypeTeacher)
}

func teamMemberView(person *model.People) TeamMemberView {
	return TeamMemberView{
		Name:   person.Name,
		Gender: person.Gender,
		Campus: person.Campus,
		ID:     int(person.ID),
		Type:   person.Type,
		Contact: TeamContact{
			QQ:     person.Qq,
			Wechat: person.Wechat,
			Tel:    person.Tel,
		},
		WalkStatus: person.WalkStatus,
	}
}

func hfTeamCreate(ctx *gin.Context) {
	api := &TeamCreateApi{}
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
