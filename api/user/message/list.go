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

func ListMessageHandler() gin.HandlerFunc {
	api := ListMessageApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfListMessage).Pointer()).Name()] = api
	return hfListMessage
}

type ListMessageApi struct {
	Info     struct{} `name:"获取消息列表" desc:"获取当前用户接收的消息列表"`
	Request  struct{}
	Response ListMessageApiResponse
}

type ListMessageApiResponse struct {
	Messages []MessageView `json:"messages" desc:"消息列表"`
}

type MessageView struct {
	ID         int64  `json:"id" desc:"消息ID"`
	SenderID   *int64 `json:"sender_id" desc:"发送者人员ID，为空表示系统消息"`
	ReceiverID int64  `json:"receiver_id" desc:"接收者人员ID"`
	Message    string `json:"message" desc:"消息内容"`
}

func (h *ListMessageApi) Init(ctx *gin.Context) error { return nil }
func (h *ListMessageApi) Run(ctx *gin.Context) kit.Code { return comm.CodeOK }

func hfListMessage(ctx *gin.Context) {
	api := &ListMessageApi{}
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
