package stats

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
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

// Run Api业务逻辑执行点
func (r *RouteApi) Run(ctx *gin.Context) kit.Code {
	// TODO: 在此处编写接口业务逻辑
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
