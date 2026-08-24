package user

import (
	"reflect"
	"runtime"

	peopleCache "app/dao/cache/people"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func UserModifyHandler() gin.HandlerFunc {
	api := UserModifyApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUserModify).Pointer()).Name()] = api
	return hfUserModify
}

type UserModifyApi struct {
	Info     struct{} `name:"修改用户信息"`
	Request  UserModifyApiRequest
	Response struct{}
}

type UserModifyApiRequest struct {
	Body struct {
		Identity string      `json:"identity" `
		Contact  UserContact `json:"contact" binding:"required"`
	}
}

func (h *UserModifyApi) Init(ctx *gin.Context) error { return ctx.ShouldBindJSON(&h.Request.Body) }
func (h *UserModifyApi) Run(ctx *gin.Context) kit.Code {
	if code := comm.CheckBizPhase(comm.PhaseRegistration, comm.PhaseSubmission, comm.PhaseAdjustment); code != comm.CodeOK {
		return code
	}
	person, code := currentUserPerson(ctx)
	if code != comm.CodeOK {
		return code
	}

	if h.Request.Body.Identity != "" {
		identity, err := comm.EncryptIdentity(h.Request.Body.Identity)
		if err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("加密身份证号失败")
			return comm.CodeServerError
		}
		existing, err := repo.NewPeopleRepo().FindPeopleByStoredIdentity(ctx, identity)
		if err != nil {
			return comm.CodeServerError
		}
		if existing != nil && existing.ID != person.ID {
			return comm.CodeAlreadyRegistered
		}
		person.Identity = identity
	}
	person.Qq = h.Request.Body.Contact.QQ
	person.Wechat = h.Request.Body.Contact.Wechat
	person.Tel = h.Request.Body.Contact.Tel

	updates := map[string]any{
		"qq":     person.Qq,
		"wechat": person.Wechat,
		"tel":    person.Tel,
	}
	if h.Request.Body.Identity != "" {
		updates["identity"] = person.Identity
	}

	if err := repo.NewPeopleRepo().UpdateByID(ctx, person.ID, updates); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("更新当前用户失败")
		return comm.CodeServerError
	}
	_ = peopleCache.DelPersonByID(ctx, person.ID)
	return comm.CodeOK
}

func hfUserModify(ctx *gin.Context) {
	api := &UserModifyApi{}
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
