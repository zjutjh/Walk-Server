package message

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

func DeleteMessageHandler() gin.HandlerFunc {
	api := DeleteMessageApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfDeleteMessage).Pointer()).Name()] = api
	return hfDeleteMessage
}

type DeleteMessageApi struct {
	Info     struct{} `name:"删除消息" desc:"删除当前用户收到的指定消息"`
	Request  DeleteMessageApiRequest
	Response struct{}
}

type DeleteMessageApiRequest struct {
	Body struct {
		MessageID int64 `json:"message_id" desc:"消息ID" binding:"required"`
	}
}

func (h *DeleteMessageApi) Init(ctx *gin.Context) error { return ctx.ShouldBindJSON(&h.Request.Body) }
func (h *DeleteMessageApi) Run(ctx *gin.Context) kit.Code { return comm.CodeOK }

func hfDeleteMessage(ctx *gin.Context) {
	api := &DeleteMessageApi{}
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
