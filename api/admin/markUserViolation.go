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
	peopleCache "app/dao/cache/people"
	repo "app/dao/repo"
)

func MarkUserViolationHandler() gin.HandlerFunc {
	api := MarkUserViolationApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(markUserViolation).Pointer()).Name()] = api
	return markUserViolation
}

type MarkUserViolationApi struct {
	Info     struct{} `name:"标记成员违规" desc:"标记或取消标记成员违规状态"`
	Request  MarkUserViolationApiRequest
	Response MarkUserViolationApiResponse
}

type MarkUserViolationApiRequest struct {
	Body struct {
		UserID     int  `json:"user_id" desc:"用户编号" binding:"required"`
		IsViolated bool `json:"is_violated" desc:"是否违规"`
	}
}

type MarkUserViolationApiResponse struct{}

func (m *MarkUserViolationApi) Run(ctx *gin.Context) kit.Code {
	peopleRepo := repo.NewPeopleRepo()

	person, err := peopleRepo.FindPeopleByID(ctx, int64(m.Request.Body.UserID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询人员失败")
		return comm.CodeServerError
	}
	if person == nil {
		return comm.CodePeopleNotFound
	}

	if err := peopleRepo.UpdateViolationByUserID(ctx, person.ID, m.Request.Body.IsViolated); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("标记成员违规失败")
		return comm.CodeServerError
	}
	_ = peopleCache.DelPersonByID(ctx, person.ID)
	return comm.CodeOK
}

func (m *MarkUserViolationApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&m.Request.Body)
}

func markUserViolation(ctx *gin.Context) {
	api := &MarkUserViolationApi{}
	err := api.Init(ctx)
	if err != nil {
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
