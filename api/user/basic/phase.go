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
	Info     struct{} `name:"获取各业务阶段时间" desc:"返回所有业务阶段的开始和结束时间"`
	Request  struct{}
	Response PhaseApiResponse
}

type PhaseApiResponse struct {
	Phases []PhaseTimeItem `json:"phases" desc:"各业务阶段的起止时间，按业务顺序排列"`
}

type PhaseTimeItem struct {
	Phase comm.BizPhase `json:"phase" desc:"业务阶段枚举值"`
	Start string        `json:"start" desc:"开始时间，格式为 YYYY-MM-DD HH:mm:ss"`
	End   string        `json:"end" desc:"结束时间，格式为 YYYY-MM-DD HH:mm:ss"`
}

func (h *PhaseApi) Run(_ *gin.Context) kit.Code {
	phases := comm.BizConf.Phases
	h.Response.Phases = []PhaseTimeItem{
		{Phase: comm.PhaseRegistration, Start: phases.Registration.Start, End: phases.Registration.End},
		{Phase: comm.PhaseSubmission, Start: phases.Submission.Start, End: phases.Submission.End},
		{Phase: comm.PhaseAdjustment, Start: phases.Adjustment.Start, End: phases.Adjustment.End},
		{Phase: comm.PhasePreparation, Start: phases.Preparation.Start, End: phases.Preparation.End},
		{Phase: comm.PhaseActivity, Start: phases.Activity.Start, End: phases.Activity.End},
	}
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
