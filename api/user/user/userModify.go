package user

import (
	"reflect"
	"runtime"

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
		Campus   string      `json:"campus"`
		College  string      `json:"college" binding:"required"`
		Identity string      `json:"identity" `
		Contact  UserContact `json:"contact" binding:"required"`
	}
}

func (h *UserModifyApi) Init(ctx *gin.Context) error { return ctx.ShouldBindJSON(&h.Request.Body) }
func (h *UserModifyApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentUserPerson(ctx)
	if code != comm.CodeOK {
		return code
	}

	if h.Request.Body.Campus != "" {
		campus, ok := comm.ParseCampus(h.Request.Body.Campus)
		if !ok {
			return comm.CodeParameterInvalid
		}
		person.Campus = campus
	}
	if h.Request.Body.Identity != "" {
		person.Identity = h.Request.Body.Identity
	}
	person.College = h.Request.Body.College
	person.Qq = h.Request.Body.Contact.QQ
	person.Wechat = h.Request.Body.Contact.Wechat
	person.Tel = h.Request.Body.Contact.Tel

	updates := map[string]any{
		"campus":  person.Campus,
		"college": person.College,
		"qq":      person.Qq,
		"wechat":  person.Wechat,
		"tel":     person.Tel,
	}
	if h.Request.Body.Identity != "" {
		updates["identity"] = person.Identity
	}

	if err := repo.NewPeopleRepo().UpdateByOpenID(ctx, person.OpenID, updates); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("更新当前用户失败")
		return comm.CodeServerError
	}
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
