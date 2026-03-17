package api

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/session"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/repo"
)

func AuthAdminHandler() gin.HandlerFunc {
    api := AuthAdminApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(authAdmin).Pointer()).Name()] = api
    return authAdmin
}

type AuthAdminApi struct {
    Info     struct{} `name:"管理员登录"`
    Request  AuthAdminApiRequest
    Response AuthAdminApiResponse
}

type AuthAdminApiRequest struct {
    Body struct{
        Account string `json:"account" desc:"管理员账号" binding:"required"`
        Password string `json:"password" desc:"密码" binding:"required"`
    } 
}

type AuthAdminApiResponse struct {
	PointName string `json:"point_name" desc:"点位名称"`
	Name string `json:"name" desc:"管理员姓名"`
}

// Run Api业务逻辑执行点
func (a *AuthAdminApi) Run(ctx *gin.Context) kit.Code {
    adminRepo := repo.NewAdminRepo()

    account := strings.TrimSpace(a.Request.Body.Account)
    rawPassword := a.Request.Body.Password

    admin, err := adminRepo.FindByAccount(ctx, account)
    if err != nil {
        nlog.Pick().WithContext(ctx).WithError(err).Error("查询管理员失败")
        return comm.CodeUnknownError
    }
    if admin == nil {
        return comm.CodeAccountOrPasswordError // 你项目里如果没这个码，就换成你自己的“账号或密码错误”
    }

    // 校验密码
    if !comm.Verify(admin.Password, rawPassword) {
        return comm.CodeAccountOrPasswordError
    }
    err=session.SetIdentity(ctx,admin.ID)
    

    a.Response = AuthAdminApiResponse{
        PointName:  admin.PointName,
        Name:       admin.Name,
    }

    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (a *AuthAdminApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&a.Request.Body)
	if err != nil {
		return err
	}
    return err
}

func authAdmin(ctx *gin.Context) {
    api := &AuthAdminApi{}
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