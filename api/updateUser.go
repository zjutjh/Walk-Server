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

func UpdateUserHandler() gin.HandlerFunc {
    api := UpdateUserApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(updateUser).Pointer()).Name()] = api
    return updateUser
}

type UpdateUserApi struct {
    Info     struct{} `name:"更改人员状态"`
    Request  UpdateUserApiRequest
    Response UpdateUserApiResponse
}

type UpdateUserApiRequest struct {
    Body struct{
        UserID int `json:"user_id"`
        Status string `json:"status" desc:"未开始,待出发,已放弃,进行中,已下撤,已违规,已完成"`
    }
}

type UpdateUserApiResponse struct {
}


func (h *UpdateUserApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}


func (h *UpdateUserApi) Init(ctx *gin.Context) (err error) {
    return err
}


func updateUser(ctx *gin.Context) {
    api := &UpdateUserApi{}
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