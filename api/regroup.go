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
        Members  []int `json:"members" desc:"用户编号,长度3-6人" binding:"required"`
	    RouteName string `json:"route_name" desc:"路线名称" binding:"required"`
    }
}

type RegroupApiResponse struct {
    TeamID int `json:"team_id" desc:"队伍编号"`
}


// Run Api业务逻辑执行点
func (r *RegroupApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (r *RegroupApi) Init(ctx *gin.Context) (err error) {
    err = ctx.ShouldBindJSON(&r.Request.Body)
	if err != nil {
		return err
	}
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