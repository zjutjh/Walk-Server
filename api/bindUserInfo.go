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

func BindUserInfoHandler() gin.HandlerFunc {
    api := BindUserInfoApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(bindUserInfo).Pointer()).Name()] = api
    return bindUserInfo
}

type BindUserInfoApi struct {
    Info     struct{} `name:"添加单人信息"`
    Request  BindUserInfoApiRequest
    Response BindUserInfoApiResponse
}

type BindUserInfoApiRequest struct {
    Body struct{
        Name string `json:"name"`
        Tel string `json:"tel"`
	    StuID string `json:"stu_id"`
    }
}

type BindUserInfoApiResponse struct {
}

// Run Api业务逻辑执行点
func (h *BindUserInfoApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (h *BindUserInfoApi) Init(ctx *gin.Context) (err error) {
    return err
}

func bindUserInfo(ctx *gin.Context) {
    api := &BindUserInfoApi{}
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