package api

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	teamCache "app/dao/cache/team"
	"app/dao/query"
	repo "app/dao/repo"
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
	teamID := int64(c.Request.Body.TeamID)
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)

		if err := txPeopleRepo.UpdateMembersWalkStatusByCurrent(ctx, teamID, comm.WalkStatusInProgress, comm.WalkStatusCompleted); err != nil {
			return err
		}
		return txTeamRepo.UpdateByID(ctx, teamID, map[string]any{"status": comm.TeamStatusCompleted})
	})
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("终点确认失败")
		return comm.CodeServerError
	}
	_ = teamCache.DelTeamByID(ctx, teamID)
	_ = teamCache.DeleteTeamInfo(ctx, teamID)
	return comm.CodeOK
}

func (c *ConfirmDestinationApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&c.Request.Body)
}

func confirmDestination(ctx *gin.Context) {
	api := &ConfirmDestinationApi{}
	if err := api.Init(ctx); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
