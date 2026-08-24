package api

import (
	"errors"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"gorm.io/gorm"

	"app/comm"
	"app/dao/query"
	repo "app/dao/repo"
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
	teamID := int64(m.Request.Body.TeamID)
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)

		if _, err := txTeamRepo.GetTeamByID(ctx, teamID); err != nil {
			return err
		}

		return txPeopleRepo.UpdateMembersViolationExceptStatuses(ctx, teamID, []string{
			comm.WalkStatusAbandoned,
			comm.WalkStatusWithdrawn,
		}, true)
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return comm.CodeTeamNotFound
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("标记队伍违规失败")
		return comm.CodeServerError
	}
	return comm.CodeOK
}

func (m *MarkTeamViolationApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&m.Request.Body)
}

func markTeamViolation(ctx *gin.Context) {
	api := &MarkTeamViolationApi{}
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
