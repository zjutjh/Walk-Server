package api

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

func SubmitTeamHandler() gin.HandlerFunc {
    api := SubmitTeamApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(submitTeam).Pointer()).Name()] = api
    return submitTeam
}
type SubmitTeamApi struct {
    Info     struct{} `name:"直接提交队伍"`
    Request  SubmitTeamApiRequest
    Response SubmitTeamApiResponse
}

type SubmitTeamApiRequest struct {
    Body struct{
        TeamID int `json:"team_id"`
	    Secret string `json:"secret"`
    }
}

type SubmitTeamApiResponse struct {
}


func (h *SubmitTeamApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}


func (h *SubmitTeamApi) Init(ctx *gin.Context) (err error) {
    return err
}


func submitTeam(ctx *gin.Context) {
    api := &SubmitTeamApi{}
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