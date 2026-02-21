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

func BindUserCodeHandler() gin.HandlerFunc {
    api := BindUserCodeApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(bindUserCode).Pointer()).Name()] = api
    return bindUserCode
}

type BindUserCodeApi struct {
    Info     struct{} `name:"扫码添加单人信息"`
    Request  BindUserCodeApiRequest
    Response BindUserCodeApiResponse
}

type BindUserCodeApiRequest struct {
    Body struct{
        Content string `json:"content" desc:"个人码"`
    }
}

type BindUserCodeApiResponse struct {
	Name string `json:"name"`
}

// Run Api业务逻辑执行点
func (h *BindUserCodeApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (h *BindUserCodeApi) Init(ctx *gin.Context) (err error) {
    return err
}

func bindUserCode(ctx *gin.Context) {
    api := &BindUserCodeApi{}
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