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

func ConfirmCheckPointHandler() gin.HandlerFunc {
    api := ConfirmCheckPointApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(confirmCheckPoint).Pointer()).Name()] = api
    return confirmCheckPoint
}

type ConfirmCheckPointApi struct {
    Info     struct{} `name:"中间点位确认"`
    Request  ConfirmCheckPointApiRequest
    Response ConfirmCheckPointApiResponse
}

type ConfirmCheckPointApiRequest struct {
    Body struct{
        TeamID int `json:"team_id"`
        Status string `json:"content" desc:"0同意1反对"`
    }
}

type ConfirmCheckPointApiResponse struct {
}


func (h *ConfirmCheckPointApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}


func (h *ConfirmCheckPointApi) Init(ctx *gin.Context) (err error) {
    return err
}


func confirmCheckPoint(ctx *gin.Context) {
    api := &ConfirmCheckPointApi{}
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