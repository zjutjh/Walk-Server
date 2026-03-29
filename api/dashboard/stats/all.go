package stats

import (
	"encoding/json"
	"reflect"
	"runtime"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	routeCache "app/dao/cache/route"
	repo "app/dao/repo"
)

// AllHandler API router注册点
func AllHandler() gin.HandlerFunc {
	api := AllApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfAll).Pointer()).Name()] = api
	return hfAll
}

type RouteStatItem struct {
	Started    int `json:"started" desc:"已出发人数"`
	NotPresent int `json:"not_present" desc:"未到场人数"`
	UnDeparted int `json:"undeparted" desc:"待出发人数"`
	TotalReg   int `json:"total_reg" desc:"总报名人数"`
	Finished   int `json:"finished" desc:"已结束人数（无论是否违规）"`
	WrongRoute int `json:"wrong_route" desc:"走错路线人数（走到另一条线路的人数）"`
	Withdrawn  int `json:"withdrawn" desc:"下撤人数"`
}

type RouteStats struct {
	RouteName string        `json:"route_name" desc:"路线代号"`
	Stats     RouteStatItem `json:"stats" desc:"统计数据"`
}

type AllApi struct {
	Info     struct{}       `name:"获取所有路线统计数据" desc:"获取所有路线的统计数据表格"`
	Request  AllApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response AllApiResponse // API响应数据 (Body中的Data部分)
}

type AllApiRequest struct {
}

type AllApiResponse struct {
	Routes []RouteStats `json:"routes" desc:"路线统计列表"`
}

func ensureRouteStat(routeStats map[string]*RouteStatItem, routeOrder *[]string, routeName string) *RouteStatItem {
	// 已存在则复用，避免重复初始化统计对象。
	stat, ok := routeStats[routeName]
	if ok {
		return stat
	}

	// 新路线首次出现时，初始化并记录到输出顺序中。
	stat = &RouteStatItem{}
	routeStats[routeName] = stat
	*routeOrder = append(*routeOrder, routeName)
	return stat
}

func applyStatus(stat *RouteStatItem, walkStatus string, count int) {
	// Map walk_status to frontend display fields.
	switch walkStatus {
	case "notStart", "abandoned":
		stat.NotPresent += count
	case "pending":
		stat.UnDeparted += count
	case "inProgress":
		stat.Started += count
	case "withdrawn":
		stat.Withdrawn += count
	case "completed", "violated":
		stat.Finished += count
	}
}

// Run Api业务逻辑执行点
func (a *AllApi) Run(ctx *gin.Context) kit.Code {
	// Redis缓存规划:
	// Key: dashboard:stats:route:all
	// Type: String(JSON)
	// TTL: 15s
	// 1) 使用 cache dao 先尝试读取 Redis。
	// 先走缓存，命中后直接返回，降低统计查询压力。
	cached, found, err := routeCache.GetAllRouteStats(ctx)
	if err != nil {
		// 非未命中错误时仅告警，继续回源数据库。
		nlog.Pick().WithContext(ctx).WithError(err).Warn("读取路线统计缓存失败")
	} else if found {
		cachedResp := AllApiResponse{}
		err = json.Unmarshal(cached, &cachedResp)
		if err == nil {
			a.Response = cachedResp
			return comm.CodeOK
		}

		nlog.Pick().WithContext(ctx).WithError(err).Warn("解析路线统计缓存失败")
	}

	// 2) 缓存未命中或异常时，回源数据库做聚合计算。
	routeRepo := repo.NewRouteRepo()

	// 2.1) 先查启用路线，保证没有报名数据的路线也能返回 0 统计。
	routes, err := routeRepo.ListActiveRouteNames(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线列表失败")
		return comm.CodeDatabaseError
	}

	// 2.2) 初始化输出顺序和统计容器。
	routeOrder := make([]string, 0, len(routes))
	routeStats := make(map[string]*RouteStatItem, len(routes))
	for _, route := range routes {
		routeOrder = append(routeOrder, route.Name)
		routeStats[route.Name] = &RouteStatItem{}
	}

	// 2.3) 查询路线+人员状态聚合，得到总报名与各状态人数。
	statusRows, err := routeRepo.ListRouteStatusCounts(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线状态统计失败")
		return comm.CodeDatabaseError
	}

	// 2.4) 将聚合行写入每条路线统计结构。
	for _, row := range statusRows {
		stat := ensureRouteStat(routeStats, &routeOrder, row.RouteName)
		count := int(row.Count)
		stat.TotalReg += count
		applyStatus(stat, row.WalkStatus, count)
	}

	// 2.5) 走错人数单独聚合，避免与人员状态口径混淆。
	wrongRows, err := routeRepo.ListRouteWrongCounts(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线走错统计失败")
		return comm.CodeDatabaseError
	}

	// 2.6) 回填每条路线的走错人数。
	for _, row := range wrongRows {
		stat := ensureRouteStat(routeStats, &routeOrder, row.RouteName)
		stat.WrongRoute = int(row.Count)
	}

	// 没有启用路线时兜底排序，保证输出顺序稳定。
	if len(routes) == 0 {
		sort.Strings(routeOrder)
	}

	// 2.7) 组装最终响应。
	a.Response.Routes = make([]RouteStats, 0, len(routeOrder))
	for _, routeName := range routeOrder {
		a.Response.Routes = append(a.Response.Routes, RouteStats{
			RouteName: routeName,
			Stats:     *routeStats[routeName],
		})
	}

	cacheBody, err := json.Marshal(a.Response)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("序列化路线统计缓存失败")
		return comm.CodeOK
	}

	// 3) 回填短 TTL 缓存，后续请求直接命中缓存。
	err = routeCache.SetAllRouteStats(ctx, cacheBody)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("写入路线统计缓存失败")
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (a *AllApi) Init(ctx *gin.Context) (err error) {
	return err
}

// hfAll API执行入口
func hfAll(ctx *gin.Context) {
	api := &AllApi{}
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
