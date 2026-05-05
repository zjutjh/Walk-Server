package api

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/session"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func LogoutAdminHandler() gin.HandlerFunc {
	api := LogoutAdminApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(logoutAdmin).Pointer()).Name()] = api
	return logoutAdmin
}

type LogoutAdminApi struct {
	Info     struct{} `name:"管理员退出登录"`
	Request  struct{}
	Response struct{}
}

func (a *LogoutAdminApi) Run(ctx *gin.Context) kit.Code {
	if err := session.DeleteIdentity(ctx); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("删除管理员登录态失败")
		return comm.CodeMiddlewareServiceError
	}
	return comm.CodeOK
}

func (a *LogoutAdminApi) Init(ctx *gin.Context) (err error) {
	return nil
}

func logoutAdmin(ctx *gin.Context) {
	api := &LogoutAdminApi{}
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
