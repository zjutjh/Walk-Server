package dashboard

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
	RouteId    string `json:"route_id" desc:"路线代号 (基于望舒文档，如 1 或者 pf-half"`
	RegNum     int    `json:"reg_num" desc:"该路线报名人数"`
	InProgress int    `json:"in_progress" desc:"该路线进行中人数"`
	NotPresent int    `json:"not_present" desc:"未到场人数"`
	WrongRoute int    `json:"wrong_route" desc:"走错路线人数"`
}

// Run Api业务逻辑执行点
func (o *OverviewApi) Run(ctx *gin.Context) kit.Code {
	// TODO: 在此处编写接口业务逻辑
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
			reply.Success(ctx, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
