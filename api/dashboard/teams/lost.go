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

// LostHandler API router注册点
func LostHandler() gin.HandlerFunc {
	api := LostApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfLost).Pointer()).Name()] = api
	return hfLost
}

type LostApi struct {
	Info     struct{}        `name:"设置队伍失联状态" desc:"设置指定队伍的失联状态 \n 距现在5min内打卡的队伍不允许设置失联状态为true"`
	Request  LostApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response LostApiResponse // API响应数据 (Body中的Data部分)
}

type LostApiRequest struct {
	Body struct {
		TeamId string `json:"team_id" desc:"队伍ID"`
		IsLost bool   `json:"is_lost" desc:"是否失联"`
	}
}

type LostApiResponse struct{}

// Run Api业务逻辑执行点
func (l *LostApi) Run(ctx *gin.Context) kit.Code {
	// TODO: 在此处编写接口业务逻辑
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (l *LostApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&l.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// hfLost API执行入口
func hfLost(ctx *gin.Context) {
	api := &LostApi{}
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
