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

func GetPointInfoHandler() gin.HandlerFunc {
    api := GetPointInfoApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(getPointInfo).Pointer()).Name()] = api
    return getPointInfo
}

type GetPointInfoApi struct {
    Info     struct{} `name:"获取点位信息"`
    Request  GetPointInfoApiRequest
    Response GetPointInfoApiResponse
}

type GetPointInfoApiRequest struct {
}

type GetPointInfoApiResponse struct {
    AdminID int `json:"admin_id"`
	RouteCode string `json:"route_code"`
	Point string `json:"point"`
}

// Run Api业务逻辑执行点
func (h *GetPointInfoApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (h *GetPointInfoApi) Init(ctx *gin.Context) (err error) {
    return err
}

func getPointInfo(ctx *gin.Context) {
    api := &GetPointInfoApi{}
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