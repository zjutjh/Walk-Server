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

func RegisterStudentHandler() gin.HandlerFunc {
	api := RegisterStudentApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterStudent).Pointer()).Name()] = api
	return hfRegisterStudent
}

type RegisterStudentApi struct {
	Info     struct{} `name:"学生报名" desc:"学生报名接口"`
	Request  RegisterCommonRequest
	Response struct{}
}

func (h *RegisterStudentApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *RegisterStudentApi) Run(ctx *gin.Context) kit.Code {
	return doRegister(ctx, h.Request, comm.MemberTypeStudent)
}

func hfRegisterStudent(ctx *gin.Context) {
	api := &RegisterStudentApi{}
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
