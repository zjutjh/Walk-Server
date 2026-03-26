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
	repo "app/dao/repo/admin"
)

func ConfirmDestinationHandler() gin.HandlerFunc {
	api := ConfirmDestinationApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(confirmDestination).Pointer()).Name()] = api
	return confirmDestination
}

type ConfirmDestinationApi struct {
	Info     struct{} `name:"终点确认"`
	Request  ConfirmDestinationApiRequest
	Response ConfirmDestinationApiResponse
}

type ConfirmDestinationApiRequest struct {
	Body struct {
		TeamID int `json:"team_id" desc:"团队编号" binding:"required"`
	}
}

type ConfirmDestinationApiResponse struct {
}

func (c *ConfirmDestinationApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()

	err := teamRepo.ConfirmDestination(ctx, int64(c.Request.Body.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("终点确认失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func (c *ConfirmDestinationApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&c.Request.Body)
	if err != nil {
		return err
	}
	return err
}

func confirmDestination(ctx *gin.Context) {
	api := &ConfirmDestinationApi{}
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
