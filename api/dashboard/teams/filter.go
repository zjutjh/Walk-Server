package teams

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

// FilterHandler API router注册点
func FilterHandler() gin.HandlerFunc {
	api := FilterApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfFilter).Pointer()).Name()] = api
	return hfFilter
}

type FilterApi struct {
	Info     struct{}          `name:"筛选队伍" desc:"搜索队伍和获取指定路段上的队伍列表合并接口，按最新更新时间排序"`
	Request  FilterApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response FilterApiResponse // API响应数据 (Body中的Data部分)
}

type FilterApiRequest struct {
	Query struct {
		ToPointId   string `form:"to_point_id" desc:"结束点位id，全局唯一，不是CPn"`
		FromPointId string `form:"from_point_id" desc:"起始点位id，合流点一定要给"`
		Key         string `form:"key" desc:"搜索关键词"`
		SearchType  string `form:"search_type" desc:"搜索类型（team_id/captain_phone）"`
		Limit       int    `form:"limit" desc:"返回数量"`
		Cursor      int    `form:"cursor" desc:"指针"`
	}
}

type FilterApiResponse struct {
	SegmentRange string          `json:"segment_range" desc:"点位范围，如CP1-CP2"`
	TotalCount   int             `json:"total_count" desc:"获取到的总队伍数"`
	Teams        []TeamBriefInfo `json:"teams" desc:"队伍列表"`
}

type TeamBriefInfo struct {
	TeamId               string `json:"team_id" desc:"队伍ID"`
	CaptainPhone         string `json:"captain_phone" desc:"队长联系电话"`
	RouteId              string `json:"route_id" desc:"路线ID"`
	LatestCheckpointId   string `json:"latest_checkpoint_id" desc:"最新经过点位唯一id"`
	LatestCheckpointTime string `json:"latest_checkpoint_time" desc:"经过点位时间"`
}

// Run Api业务逻辑执行点
func (f *FilterApi) Run(ctx *gin.Context) kit.Code {
	// TODO: 在此处编写接口业务逻辑
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (f *FilterApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&f.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfFilter API执行入口
func hfFilter(ctx *gin.Context) {
	api := &FilterApi{}
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
