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

func BindCodeHandler() gin.HandlerFunc {
	api := BindCodeApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(bindCode).Pointer()).Name()] = api
	return bindCode
}

type BindCodeApi struct {
	Info     struct{} `name:"绑定签到码"`
	Request  BindCodeApiRequest
	Response BindCodeApiResponse
}

type BindCodeApiRequest struct {
	Body struct {
		TeamID  int    `json:"team_id" desc:"团队编号" binding:"required"`
		Content string `json:"content" desc:"签到码" binding:"required"`
	}
}

type BindCodeApiResponse struct {
}

const (
	minTeamMemberCount = 3
)

// Run Api业务逻辑执行点
func (b *BindCodeApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()
	peopleRepo := repo.NewPeopleRepo()

	team, err := teamRepo.FindTeamByID(ctx, int64(b.Request.Body.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍失败")
		return comm.CodeUnknownError
	}
	if team == nil {
		return comm.CodeTeamNotFound
	}

	pendingCount, err := peopleRepo.CountPendingMembers(ctx, int64(b.Request.Body.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("统计待出发人数失败")
		return comm.CodeUnknownError
	}
	if pendingCount < minTeamMemberCount {
		return comm.CodeTeamMemberInsufficient
	}

	err = teamRepo.BindCodeAndStartPendingMembers(ctx, int64(b.Request.Body.TeamID), b.Request.Body.Content)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("绑定签到码失败")
		return comm.CodeBindCodeError
	}

	return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (b *BindCodeApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&b.Request.Body)
	if err != nil {
		return err
	}
	return err
}

func bindCode(ctx *gin.Context) {
	api := &BindCodeApi{}
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
