package api

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/model"
	repo "app/dao/repo"
)

func GetTeamStatusHandler() gin.HandlerFunc {
	api := GetTeamStatusApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(getTeamStatus).Pointer()).Name()] = api
	return getTeamStatus
}

type GetTeamStatusApi struct {
	Info     struct{} `name:"获取团队状态"`
	Request  GetTeamStatusApiRequest
	Response GetTeamStatusApiResponse
}

type GetTeamStatusApiRequest struct {
	Query struct {
		TeamID int `form:"team_id" binding:"required"`
	}
}

type GetTeamStatusApiResponse struct {
	Team    TeamResponse     `json:"team" `
	Members []MemberResponse `json:"members"`
}

type TeamResponse struct {
	Name                  string `json:"name" desc:"队名"`
	PrevPointName         string `json:"prev_point_name" desc:"前序点位名称"`
	LatestPointName       string `json:"latest_point_name" desc:"最新经过点位名称"`
	RouteName             string `json:"route_name" desc:"路线名称"`
	Status                string `json:"status" desc:"队伍状态"`
	IsWrongRoute          bool   `json:"is_wrong_route" desc:"是否走错路线"`
	IsPrevPointInvalid    bool   `json:"is_prev_point_invalid" desc:"前序点位是否违反路线顺序"`
	IsJustEnterWrongRoute bool   `json:"is_just_enter_wrong_route" desc:"最后一次打卡是否导致队伍进入错路状态"`
}

type MemberResponse struct {
	Name       string `json:"name" desc:"姓名"`
	UserID     int    `json:"user_id" desc:"用户编号"`
	WalkStatus string `json:"walk_status" desc:"用户状态"`
	IsViolated bool   `json:"is_violated" desc:"是否违规"`
	Role       string `json:"role" desc:"用户身份"`
}

// Run Api业务逻辑执行点
func (g *GetTeamStatusApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()
	peopleRepo := repo.NewPeopleRepo()

	team, err := teamRepo.FindTeamByID(ctx, int64(g.Request.Query.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍状态失败")
		return comm.CodeServerError
	}
	if team == nil {
		return comm.CodeTeamNotFound
	}

	routeStatus, err := g.resolveRouteStatus(ctx, teamRepo, team)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("计算队伍路线状态失败")
		return comm.CodeServerError
	}

	members, err := peopleRepo.FindPeopleByTeamID(ctx, int64(g.Request.Query.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍成员失败")
		return comm.CodeServerError
	}

	g.Response.Team = TeamResponse{
		Name:            team.Name,
		PrevPointName:   routeStatus.prevPointName,
		LatestPointName: team.LatestPointName,
		RouteName:       team.RouteName,
		Status:          team.Status,

		IsWrongRoute:          routeStatus.isWrongRoute,
		IsPrevPointInvalid:    routeStatus.isPrevPointInvalid,
		IsJustEnterWrongRoute: routeStatus.isJustEnterWrongRoute,
	}

	g.Response.Members = make([]MemberResponse, 0, len(members))
	for _, member := range members {
		g.Response.Members = append(g.Response.Members, MemberResponse{
			Name:       member.Name,
			UserID:     int(member.ID),
			WalkStatus: member.WalkStatus,
			IsViolated: member.IsViolated,
			Role:       member.Role,
		})
	}

	return comm.CodeOK
}

type teamRouteStatus struct {
	isWrongRoute          bool
	prevPointName         string
	isPrevPointInvalid    bool
	isJustEnterWrongRoute bool
}

func (g *GetTeamStatusApi) resolveRouteStatus(ctx *gin.Context, teamRepo *repo.TeamRepo, team *model.Team) (teamRouteStatus, error) {
	status := teamRouteStatus{
		isWrongRoute: team.IsWrongRoute,
	}

	checkins, err := teamRepo.ListLatestCheckins(ctx, team.ID, 2)
	if err != nil {
		return status, err
	}
	if len(checkins) == 0 {
		return status, nil
	}

	activeRouteName := team.RouteName
	if status.isWrongRoute {
		wrongRouteName, found, err := teamRepo.GetLatestWrongRouteName(ctx, team.ID)
		if err != nil {
			return status, err
		}
		if found {
			activeRouteName = wrongRouteName
		}
	}

	latestCheckin := checkins[0]
	if len(checkins) >= 2 {
		prevCheckin := checkins[1]
		status.prevPointName = prevCheckin.PointName
		isValid, err := teamRepo.IsRouteTransitionValid(ctx, activeRouteName, prevCheckin.PointName, latestCheckin.PointName)
		if err != nil {
			return status, err
		}
		status.isPrevPointInvalid = !isValid
	}

	latestOnOriginalRoute, err := teamRepo.IsPointOnRoute(ctx, team.RouteName, latestCheckin.PointName)
	if err != nil {
		return status, err
	}
	if latestOnOriginalRoute {
		return status, nil
	}
	if len(checkins) == 1 {
		status.isJustEnterWrongRoute = true
		return status, nil
	}

	prevOnOriginalRoute, err := teamRepo.IsPointOnRoute(ctx, team.RouteName, checkins[1].PointName)
	if err != nil {
		return status, err
	}
	status.isJustEnterWrongRoute = prevOnOriginalRoute
	return status, nil
}

// Run Api初始化 进行参数校验和绑定
func (g *GetTeamStatusApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindQuery(&g.Request.Query)
}

func getTeamStatus(ctx *gin.Context) {
	api := &GetTeamStatusApi{}
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
