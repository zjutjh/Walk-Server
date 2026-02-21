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

func RegroupHandler() gin.HandlerFunc {
    api := RegroupApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(regroup).Pointer()).Name()] = api
    return regroup
}

type RegroupApi struct {
    Info     struct{} `name:"重组队伍"`
    Request  RegroupApiRequest
    Response RegroupApiResponse
}

type RegroupApiRequest struct {
    Body struct{
        Jwts  []string `json:"jwts"`
	    RouteCode string `json:"route_code"`
	    Name string `json:"name"`
	    Slogan string `json:"slogan"`
	    Secret string `json:"secret"`
    }
}

type RegroupApiResponse struct {
}

// Run Api业务逻辑执行点
func (h *RegroupApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (h *RegroupApi) Init(ctx *gin.Context) (err error) {
    return err
}

func regroup(ctx *gin.Context) {
    api := &RegroupApi{}
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