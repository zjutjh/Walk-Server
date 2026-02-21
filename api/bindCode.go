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

func BindCodeHandler() gin.HandlerFunc {
    api := BindCodeApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(bindCode).Pointer()).Name()] = api
    return bindCode
}

type BindCodeApi struct {
    Info     struct{} `name:"绑定签到码"`
    Request  BindCodeApiRequest
    Response BindCodeApiResponse
}

type BindCodeApiRequest struct {
    Body struct{
        TeamID int `json:"team_id"`
        Content string `json:"content"`
    } 
}

type BindCodeApiResponse struct {
}

// Run Api业务逻辑执行点
func (h *BindCodeApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (h *BindCodeApi) Init(ctx *gin.Context) (err error) {
    return err
}

func bindCode(ctx *gin.Context) {
    api := &BindCodeApi{}
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