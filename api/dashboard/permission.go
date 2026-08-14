package dashboard

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	repo "app/dao/repo"
)

// PermissionHandler API router注册点
func PermissionHandler() gin.HandlerFunc {
	api := PermissionApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfPermission).Pointer()).Name()] = api
	return hfPermission
}

type PermissionApi struct {
	Info     struct{}              `name:"获取当前管理员权限信息" desc:"获取当前登录管理员的权限级别"`
	Request  PermissionApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response PermissionApiResponse // API响应数据 (Body中的Data部分)
}

type PermissionApiRequest struct {
}

type PermissionApiResponse struct {
	Name       string `json:"name" desc:"管理员姓名"`
	Permission string `json:"permission" desc:"权限级别(super最高权限,manager负责人权限,internal内部权限,external外部权限)"`
	Campus     string `json:"campus" desc:"负责校区"`
	PointName  string `json:"point_name" desc:"负责的点位name"`
}

func (p *PermissionApi) Run(ctx *gin.Context) kit.Code {
	admin, ok := repo.GetAdminInfo(ctx)
	if !ok {
		return comm.CodeServerError
	}

	p.Response.Name = admin.Name
	p.Response.Permission = admin.Permission
	p.Response.Campus = admin.Campus
	p.Response.PointName = admin.PointName

	return comm.CodeOK
}

func (p *PermissionApi) Init(ctx *gin.Context) (err error) {
	return err
}

func hfPermission(ctx *gin.Context) {
	api := &PermissionApi{}
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
