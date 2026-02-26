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

// SegmentHandler API router注册点
func SegmentHandler() gin.HandlerFunc {
	api := SegmentApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfSegment).Pointer()).Name()] = api
	return hfSegment
}

type SegmentApi struct {
	Info     struct{}           `name:"获取路段（边）信息" desc:"获取指定路段的人数信息"`
	Request  SegmentApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response SegmentApiResponse // API响应数据 (Body中的Data部分)
}

type SegmentApiRequest struct {
	Query struct {
		ToPointName   string `form:"to_point_name" desc:"结束点位name，全局唯一，不是CPn"`
		PrevPointName string `form:"prev_point_name" desc:"起始点位name，合流点一定要给"`
	}
}

type SegmentApiResponse struct {
	Number int `json:"number" desc:"该路段上的人数"`
}

// Run Api业务逻辑执行点
func (s *SegmentApi) Run(ctx *gin.Context) kit.Code {
	// TODO: 在此处编写接口业务逻辑
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (s *SegmentApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&s.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfSegment API执行入口
func hfSegment(ctx *gin.Context) {
	api := &SegmentApi{}
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
