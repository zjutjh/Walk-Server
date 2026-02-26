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
	Finished   int `json:"finished" desc:"已完成人数"`
	WrongRoute int `json:"wrong_route" desc:"走错路线人数"`
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

// Run Api业务逻辑执行点
func (a *AllApi) Run(ctx *gin.Context) kit.Code {
	// TODO: 在此处编写接口业务逻辑
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
