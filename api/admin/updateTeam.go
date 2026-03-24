package api

import (
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/session"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/model"
	"app/dao/repo"
)

func UpdateTeamHandler() gin.HandlerFunc {
	api := UpdateTeamApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(updateTeam).Pointer()).Name()] = api
	return updateTeam
}

type UpdateTeamApi struct {
	Info     struct{} `name:"打卡"`
	Request  UpdateTeamApiRequest
	Response UpdateTeamApiResponse
}

type UpdateTeamApiRequest struct {
	Body struct {
		CodeType string `json:"code_type" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
}

type UpdateTeamApiResponse struct {
	TeamID int `json:"team_id" desc:"队伍编号"`
}

// Run Api业务逻辑执行点
func (u *UpdateTeamApi) Run(ctx *gin.Context) kit.Code {
	adminRepo := repo.NewAdminRepo()
	teamRepo := repo.NewTeamRepo()

	adminID, err := session.GetIdentity[int64](ctx)
	if err != nil {
		adminIDInt, fallbackErr := session.GetIdentity[int](ctx)
		if fallbackErr != nil {
			nlog.Pick().WithContext(ctx).WithError(fallbackErr).Warn("获取管理员登录态失败")
			return comm.CodeNotLoggedIn
		}
		adminID = int64(adminIDInt)
	}

	admin, err := adminRepo.FindByID(ctx, adminID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询管理员失败")
		return comm.CodeUnknownError
	}
	if admin == nil {
		return comm.CodeNotLoggedIn
	}

	team, code := u.checkRoute(ctx, admin)
	if code != nil {
		return *code
	}

	businessCode, err := teamRepo.UpdateTeamCheckin(ctx, team, admin.PointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("更新队伍签到点失败")
		return comm.CodeUnknownError
	}
	if businessCode != nil {
		return *businessCode
	}

	u.Response.TeamID = int(team.ID)
	return comm.CodeOK
}

func (u *UpdateTeamApi) checkRoute(ctx *gin.Context, admin *model.Admin) (*model.Team, *kit.Code) {
	teamRepo := repo.NewTeamRepo()

	content := strings.TrimSpace(u.Request.Body.Content)
	codeType := strings.TrimSpace(u.Request.Body.CodeType)

	var (
		team *model.Team
		err  error
	)

	switch codeType {
	case comm.CodeTeam:
		teamID, parseErr := strconv.ParseInt(content, 10, 64)
		if parseErr != nil {
			return nil, &comm.CodeParameterInvalid
		}
		team, err = teamRepo.FindTeamByID(ctx, teamID)
	case comm.CodeChekin:
		team, err = teamRepo.FindByCode(ctx, content)
	default:
		return nil, &comm.CodeParameterInvalid
	}
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍失败")
		return nil, &comm.CodeUnknownError
	}
	if team == nil {
		return nil, &comm.CodeTeamNotFound
	}

	route, err := teamRepo.FindRouteByName(ctx, team.RouteName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线失败")
		return nil, &comm.CodeUnknownError
	}
	if route == nil {
		return nil, &comm.CodeDataNotFound
	}
	if route.Campus != admin.Campus {
		return nil, &comm.CodeCampusMismatch
	}

	return team, nil
}

// Run Api初始化 进行参数校验和绑定
func (u *UpdateTeamApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&u.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// updateTeam Api执行入口
func updateTeam(ctx *gin.Context) {
	api := &UpdateTeamApi{}
	err := api.Init(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Success(ctx, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
