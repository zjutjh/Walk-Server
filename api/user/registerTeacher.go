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

func RegisterTeacherHandler() gin.HandlerFunc {
	api := RegisterTeacherApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterTeacher).Pointer()).Name()] = api
	return hfRegisterTeacher
}

type RegisterTeacherApi struct {
	Info     struct{} `name:"教职工报名" desc:"教职工报名接口"`
	Request  RegisterCommonRequest
	Response struct{}
}

func (h *RegisterTeacherApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *RegisterTeacherApi) Run(ctx *gin.Context) kit.Code {
	return doRegister(ctx, h.Request, comm.MemberTypeTeacher)
}

func hfRegisterTeacher(ctx *gin.Context) {
	api := &RegisterTeacherApi{}
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
