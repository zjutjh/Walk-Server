package team

import (
	"reflect"
	"runtime"
	"sort"

	"app/comm"
	teamCache "app/dao/cache/team"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
)

func TeamQuotaHandler() gin.HandlerFunc {
	api := TeamQuotaApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamQuota).Pointer()).Name()] = api
	return hfTeamQuota
}

type TeamQuotaApi struct {
	Info     struct{} `name:"查询当天路线名额" desc:"查询当前是否可提交及各路线当天是否还有队伍名额"`
	Request  struct{}
	Response TeamQuotaApiResponse
}

type TeamQuotaApiResponse struct {
	CanSubmit bool                 `json:"can_submit" desc:"当前是否处于允许提交的时间段"`
	Routes    []TeamQuotaRouteItem `json:"routes" desc:"各路线当天名额状态"`
}

type TeamQuotaRouteItem struct {
	RouteName string `json:"route_name" desc:"路线代码"`
	Available bool   `json:"available" desc:"当前是否还有队伍名额"`
}

func (h *TeamQuotaApi) Run(ctx *gin.Context) kit.Code {
	routeNames := make([]string, 0, len(comm.BizConf.DailyTeamLimits))
	for routeName := range comm.BizConf.DailyTeamLimits {
		routeNames = append(routeNames, routeName)
	}
	sort.Strings(routeNames)
	h.Response.Routes = make([]TeamQuotaRouteItem, 0, len(routeNames))

	day, canSubmit := comm.CurrentSubmissionDay()
	h.Response.CanSubmit = canSubmit
	if !canSubmit {
		for _, routeName := range routeNames {
			h.Response.Routes = append(h.Response.Routes, TeamQuotaRouteItem{RouteName: routeName})
		}
		return comm.CodeOK
	}

	availability, err := teamCache.GetDailyTeamQuotaAvailability(ctx, routeNames, day)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询当天路线名额失败")
		return comm.CodeServerError
	}
	for _, routeName := range routeNames {
		h.Response.Routes = append(h.Response.Routes, TeamQuotaRouteItem{
			RouteName: routeName,
			Available: availability[routeName],
		})
	}
	return comm.CodeOK
}

func hfTeamQuota(ctx *gin.Context) {
	api := &TeamQuotaApi{}
	if code := api.Run(ctx); code == comm.CodeOK {
		reply.Reply(ctx, comm.CodeOK, api.Response)
	} else {
		reply.Fail(ctx, code)
	}
}
