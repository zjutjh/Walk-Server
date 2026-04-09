package stats

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

// RouteHandler API router注册点
func RouteHandler() gin.HandlerFunc {
	api := RouteApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRoute).Pointer()).Name()] = api
	return hfRoute
}

type PointStat struct {
	PointName   string `json:"point_name" desc:"点位唯一name"`
	PassedCount int    `json:"passed_count" desc:"经过该点位的总人数"`
}

type StatusStat struct {
	TotalReg    int `json:"total_reg" desc:"总报名人数"`
	UnPresented int `json:"unpresented" desc:"未到场人数"`
	Walking     int `json:"walking" desc:"进行中人数"`
	WrongRoute  int `json:"wrong_route" desc:"走错路线人数"`
	Withdrawn   int `json:"withdrawn" desc:"下撤人数"`
}

type RouteApi struct {
	Info     struct{}         `name:"获取特定路线详细统计" desc:"获取指定路线的详细统计数据"`
	Request  RouteApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response RouteApiResponse // API响应数据 (Body中的Data部分)
}

type RouteApiRequest struct {
	Query struct {
		Name string `form:"name" desc:"路线代号，如pf-half"`
	}
}

type RouteApiResponse struct {
	PointStats  []PointStat `json:"point_stats" desc:"经过点位总人数统计"`
	StatusStats StatusStat  `json:"status_stats" desc:"状态信息统计"`
}

func applyRouteStatus(stat *StatusStat, walkStatus string, count int) {
	switch walkStatus {
	case "notStart", "pending", "abandoned":
		stat.UnPresented += count
	case "inProgress":
		stat.Walking += count
	case "withdrawn":
		stat.Withdrawn += count
	}
}

// Run Api业务逻辑执行点
func (r *RouteApi) Run(ctx *gin.Context) kit.Code {
	// Redis缓存规划:
	// Key: walk:dashboard:stats:route:detail:{campus}:{routeName}
	// Type: String(JSON)
	// TTL: 15s
	routeName := strings.TrimSpace(r.Request.Query.Name)
	if routeName == "" {
		return comm.CodeParameterInvalid
	}

	// 先走缓存，命中则直接返回。
	cached, found, err := routeCache.GetRouteDetailStats(ctx, routeName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("读取单路线统计缓存失败")
	} else if found {
		cachedResp := RouteApiResponse{}
		err = json.Unmarshal(cached, &cachedResp)
		if err == nil {
			r.Response = cachedResp
			return comm.CodeOK
		}

		nlog.Pick().WithContext(ctx).WithError(err).Warn("解析单路线统计缓存失败")
	}

	routeRepo := repo.NewRouteRepo()

	exists, err := routeRepo.ExistsActiveRoute(ctx, routeName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("校验路线存在失败")
		return comm.CodeDatabaseError
	}
	if !exists {
		return comm.CodeDataNotFound
	}

	pointRows, err := routeRepo.ListRoutePoints(ctx, routeName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线点位失败")
		return comm.CodeDatabaseError
	}

	passedRows, err := routeRepo.ListRoutePointPassedCounts(ctx, routeName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询点位经过人数失败")
		return comm.CodeDatabaseError
	}

	passedMap := make(map[string]int, len(passedRows))
	for _, row := range passedRows {
		passedMap[row.PointName] = int(row.Count)
	}

	r.Response.PointStats = make([]PointStat, 0, len(pointRows)+len(passedRows))
	pointSeen := make(map[string]struct{}, len(pointRows))
	for _, row := range pointRows {
		r.Response.PointStats = append(r.Response.PointStats, PointStat{
			PointName:   row.PointName,
			PassedCount: passedMap[row.PointName],
		})
		pointSeen[row.PointName] = struct{}{}
	}

	extraPointNames := make([]string, 0)
	for _, row := range passedRows {
		if _, exists := pointSeen[row.PointName]; exists {
			continue
		}

		extraPointNames = append(extraPointNames, row.PointName)
	}
	for _, pointName := range extraPointNames {
		r.Response.PointStats = append(r.Response.PointStats, PointStat{
			PointName:   pointName,
			PassedCount: passedMap[pointName],
		})
	}

	statusRows, err := routeRepo.ListSingleRouteStatusCounts(ctx, routeName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线状态统计失败")
		return comm.CodeDatabaseError
	}

	status := StatusStat{}
	for _, row := range statusRows {
		count := int(row.Count)
		status.TotalReg += count
		applyRouteStatus(&status, row.WalkStatus, count)
	}

	wrongCount, err := routeRepo.CountSingleRouteWrongPeople(ctx, routeName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线走错统计失败")
		return comm.CodeDatabaseError
	}
	status.WrongRoute = int(wrongCount)
	r.Response.StatusStats = status

	cacheBody, err := json.Marshal(r.Response)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("序列化单路线统计缓存失败")
		return comm.CodeOK
	}

	err = routeCache.SetRouteDetailStats(ctx, routeName, cacheBody)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("写入单路线统计缓存失败")
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (r *RouteApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&r.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfRoute API执行入口
func hfRoute(ctx *gin.Context) {
	api := &RouteApi{}
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
