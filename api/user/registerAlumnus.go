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

func RegisterAlumnusHandler() gin.HandlerFunc {
	api := RegisterAlumnusApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterAlumnus).Pointer()).Name()] = api
	return hfRegisterAlumnus
}

type RegisterAlumnusApi struct {
	Info     struct{} `name:"校友报名" desc:"校友报名接口"`
	Request  RegisterCommonRequest
	Response struct{}
}

func (h *RegisterAlumnusApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *RegisterAlumnusApi) Run(ctx *gin.Context) kit.Code {
	return doRegister(ctx, h.Request, comm.MemberTypeAlumnus)
}

func hfRegisterAlumnus(ctx *gin.Context) {
	api := &RegisterAlumnusApi{}
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
