package dashboard

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	routeCache "app/dao/cache/route"
	repo "app/dao/repo"
)

// OverviewHandler API router注册点
func OverviewHandler() gin.HandlerFunc {
	api := OverviewApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfOverview).Pointer()).Name()] = api
	return hfOverview
}

type OverviewApi struct {
	Info     struct{}            `name:"获取总数据（地图展示页面）" desc:"获取数据大盘总览信息，包括：\n- 总报名人数\n- 进行中人数（各路线）\n- 走错路线人数（各路线）\n\n"`
	Request  OverviewApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response OverviewApiResponse // API响应数据 (Body中的Data部分)
}

type OverviewApiRequest struct {
	Query struct {
		Campus string `form:"campus" desc:"校区"`
	}
}

type OverviewApiResponse struct {
	Routes []RoutesRes `json:"routes"`
}

type RoutesRes struct {
	RouteName  string `json:"route_name" desc:"路线name"`
	TotalReg   int    `json:"total_reg" desc:"总报名人数"`
	Walking    int    `json:"walking" desc:"进行中人数"`
	Finished   int    `json:"finished" desc:"到达终点人数（无论是否违规）"`
	WrongRoute int    `json:"wrong_route" desc:"走错路线人数"`
}

func applyOverviewStatus(route *RoutesRes, walkStatus string, count int) {
	switch walkStatus {
	case "inProgress":
		route.Walking += count
	case "completed", "violated":
		route.Finished += count
	}
}

// Run Api业务逻辑执行点
func (o *OverviewApi) Run(ctx *gin.Context) kit.Code {
	// Redis缓存规划:
	// Key: walk:dashboard:overview:{campus}
	// Type: String(JSON)
	// TTL: 15s
	campus := strings.ToLower(strings.TrimSpace(o.Request.Query.Campus))
	if campus == "" {
		return comm.CodeParameterInvalid
	}

	// 先走缓存，命中则直接返回。
	cached, found, err := routeCache.GetOverview(ctx, campus)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("读取总览缓存失败")
	} else if found {
		cachedResp := OverviewApiResponse{}
		err = json.Unmarshal(cached, &cachedResp)
		if err == nil {
			o.Response = cachedResp
			return comm.CodeOK
		}

		nlog.Pick().WithContext(ctx).WithError(err).Warn("解析总览缓存失败")
	}

	routeRepo := repo.NewRouteRepo()

	routes, err := routeRepo.ListActiveRouteNamesByCampus(ctx, campus)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询总览路线失败")
		return comm.CodeDatabaseError
	}

	routeOrder := make([]string, 0, len(routes))
	routeStats := make(map[string]*RoutesRes, len(routes))
	for _, route := range routes {
		routeOrder = append(routeOrder, route.Name)
		routeStats[route.Name] = &RoutesRes{RouteName: route.Name}
	}

	statusRows, err := routeRepo.ListRouteStatusCountsByCampus(ctx, campus)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询总览状态统计失败")
		return comm.CodeDatabaseError
	}

	for _, row := range statusRows {
		stat, ok := routeStats[row.RouteName]
		if !ok {
			stat = &RoutesRes{RouteName: row.RouteName}
			routeStats[row.RouteName] = stat
			routeOrder = append(routeOrder, row.RouteName)
		}

		count := int(row.Count)
		stat.TotalReg += count
		applyOverviewStatus(stat, row.WalkStatus, count)
	}

	wrongRows, err := routeRepo.ListRouteWrongCountsByCampus(ctx, campus)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询总览走错统计失败")
		return comm.CodeDatabaseError
	}

	for _, row := range wrongRows {
		stat, ok := routeStats[row.RouteName]
		if !ok {
			stat = &RoutesRes{RouteName: row.RouteName}
			routeStats[row.RouteName] = stat
			routeOrder = append(routeOrder, row.RouteName)
		}

		stat.WrongRoute = int(row.Count)
	}

	o.Response.Routes = make([]RoutesRes, 0, len(routeOrder))
	for _, routeName := range routeOrder {
		o.Response.Routes = append(o.Response.Routes, *routeStats[routeName])
	}

	cacheBody, err := json.Marshal(o.Response)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("序列化总览缓存失败")
		return comm.CodeOK
	}

	err = routeCache.SetOverview(ctx, campus, cacheBody)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("写入总览缓存失败")
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (o *OverviewApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&o.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfOverview API执行入口
func hfOverview(ctx *gin.Context) {
	api := &OverviewApi{}
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
