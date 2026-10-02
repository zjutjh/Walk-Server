package basic

import (
	"reflect"
	"runtime"

	"app/comm"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/swagger"
)

func PhaseHandler() gin.HandlerFunc {
	api := PhaseApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfPhase).Pointer()).Name()] = api
	return hfPhase
}

type PhaseApi struct {
	Info     struct{} `name:"获取当前阶段" `
	Request  struct{}
	Response PhaseApiResponse
}

type PhaseApiResponse struct {
	Phase comm.BizPhase `json:"phase" desc:"业务时期：registration、submission、adjustment、preparation、activity；空字符串表示不在活动时期内"`
}

func (h *PhaseApi) Run(_ *gin.Context) kit.Code {
	h.Response.Phase = comm.CurrentBizPhase()
	return comm.CodeOK
}

func hfPhase(ctx *gin.Context) {
	api := &PhaseApi{}
	if code := api.Run(ctx); code == comm.CodeOK {
		reply.Reply(ctx, comm.CodeOK, api.Response)
	} else {
		reply.Fail(ctx, code)
	}
}
