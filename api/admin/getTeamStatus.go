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
	repo "app/dao/repo/admin"
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
	Name          string `json:"name" desc:"队名"`
	PrevPointName string `json:"prev_point_name" desc:"点位名称"`
	RouteName     string `json:"route_name" desc:"路线名称"`
}

type MemberResponse struct {
	Name       string `json:"name" desc:"姓名"`
	UserID     int    `json:"user_id" desc:"用户编号"`
	WalkStatus string `json:"walk_status" desc:"用户状态"`
	Role       string `json:"role" desc:"用户身份"`
}

// Run Api业务逻辑执行点
func (g *GetTeamStatusApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()
	peopleRepo := repo.NewPeopleRepo()

	team, err := teamRepo.FindTeamByID(ctx, int64(g.Request.Query.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍状态失败")
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeTeamNotFound
	}

	prevCheckinPointName := ""
	if team.PrevPointName != "" {
		routeEdge, findErr := teamRepo.FindRouteEdge(ctx, team.RouteName, team.PrevPointName)
		if findErr != nil {
			nlog.Pick().WithContext(ctx).WithError(findErr).Error("查询队伍上一签到点失败")
			return comm.CodeDatabaseError
		}
		if routeEdge != nil {
			prevCheckinPointName = routeEdge.PrevPointName
		}
	}

	members, err := peopleRepo.FindPeopleByTeamID(ctx, int64(g.Request.Query.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍成员失败")
		return comm.CodeDatabaseError
	}

	g.Response.Team = TeamResponse{
		Name:          team.Name,
		PrevPointName: prevCheckinPointName,
		RouteName:     team.RouteName,
	}

	g.Response.Members = make([]MemberResponse, 0, len(members))
	for _, member := range members {
		g.Response.Members = append(g.Response.Members, MemberResponse{
			Name:       member.Name,
			UserID:     int(member.ID),
			WalkStatus: member.WalkStatus,
			Role:       member.Role,
		})
	}

	return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (g *GetTeamStatusApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

func getTeamStatus(ctx *gin.Context) {
	api := &GetTeamStatusApi{}
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
