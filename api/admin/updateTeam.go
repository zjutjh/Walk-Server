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

func UpdateTeamHandler() gin.HandlerFunc {
    api := UpdateTeamApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(updateTeam).Pointer()).Name()] = api
    return updateTeam
}

type UpdateTeamApi struct {
    Info     struct{} `name:"打卡"`
    Request  UpdateTeamApiRequest
    Response UpdateTeamApiResponse
}

type UpdateTeamApiRequest struct {
    Body struct{
        CodeType string `json:"code_type" binding:"required"`
        Content string `json:"content" binding:"required"`
    }
}

type UpdateTeamApiResponse struct {
    TeamID int `json:"team_id" desc:"队伍编号"`
}

// Run Api业务逻辑执行点
func (u *UpdateTeamApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (u *UpdateTeamApi) Init(ctx *gin.Context) (err error) {
    err = ctx.ShouldBindJSON(&u.Request.Body)
	if err != nil {
		return err
	}
    return err
}

// updateTeam Api执行入口
func updateTeam(ctx *gin.Context) {
    api := &UpdateTeamApi{}
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