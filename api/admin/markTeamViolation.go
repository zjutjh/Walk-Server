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
	"app/dao/repo"
)

func MarkTeamViolationHandler() gin.HandlerFunc {
	api := MarkTeamViolationApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(markTeamViolation).Pointer()).Name()] = api
	return markTeamViolation
}

type MarkTeamViolationApi struct {
	Info     struct{} `name:"标记队伍违规"`
	Request  MarkTeamViolationApiRequest
	Response MarkTeamViolationApiResponse
}

type MarkTeamViolationApiRequest struct {
	Body struct {
		TeamID int `json:"team_id" desc:"团队编号" binding:"required"`
	}
}

type MarkTeamViolationApiResponse struct {
}

func (m *MarkTeamViolationApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()

	err := teamRepo.MarkViolation(ctx, int64(m.Request.Body.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("标记队伍违规失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func (m *MarkTeamViolationApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&m.Request.Body)
	if err != nil {
		return err
	}
	return err
}

func markTeamViolation(ctx *gin.Context) {
	api := &MarkTeamViolationApi{}
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
