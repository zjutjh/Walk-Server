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
        CodeType int `json:"code_type" desc:"0为团队码,1为签到码"`
        Content string `json:"content"`
    }
}

type UpdateTeamApiResponse struct {
    ProgressNum int `json:"progress_num" desc:"队伍剩余人数"`
}

// Run Api业务逻辑执行点
func (h *UpdateTeamApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (h *UpdateTeamApi) Init(ctx *gin.Context) (err error) {
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