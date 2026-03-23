package api

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/model"
	"app/dao/repo"
)

func RegisterAdminHandler() gin.HandlerFunc {
	api := RegisterAdminApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(registerAdmin).Pointer()).Name()] = api
	return registerAdmin
}

type RegisterAdminApi struct {
	Info     struct{} `name:"管理员注册（测试）"`
	Request  RegisterAdminApiRequest
	Response RegisterAdminApiResponse
}

type RegisterAdminApiRequest struct {
	Body struct {
		Account    string `json:"account" desc:"管理员账号" binding:"required"`
		Password   string `json:"password" desc:"密码" binding:"required"`
		Name       string `json:"name" desc:"管理员姓名"`
		PointName  string `json:"point_name" desc:"点位名称"`
		Permission string `json:"permission" desc:"权限级别"`
		Campus     string `json:"campus" desc:"负责校区"`
	}
}

type RegisterAdminApiResponse struct {
	ID int64 `json:"id" desc:"管理员ID"`
}

// Run Api业务逻辑执行点
func (a *RegisterAdminApi) Run(ctx *gin.Context) kit.Code {
	adminRepo := repo.NewAdminRepo()

	account := strings.TrimSpace(a.Request.Body.Account)
	password := strings.TrimSpace(a.Request.Body.Password)
	name := strings.TrimSpace(a.Request.Body.Name)
	pointName := strings.TrimSpace(a.Request.Body.PointName)
	permission := strings.TrimSpace(a.Request.Body.Permission)
	campus := strings.TrimSpace(a.Request.Body.Permission)
	if account == "" || password == "" {
		return comm.CodeParameterInvalid
	}

	// 默认值，便于测试
	if name == "" {
		name = "测试管理员"
	}
	if permission == "" {
		permission = "internal"
	}

	// 1. 检查账号是否已存在
	existAdmin, err := adminRepo.FindByAccount(ctx, account)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询管理员账号失败")
		return comm.CodeUnknownError
	}
	if existAdmin != nil {
		return comm.CodeAccountExistError
	}

	// 2. 密码 bcrypt 哈希
	hashedPassword, err := comm.Hash(password)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("管理员密码加密失败")
		return comm.CodeUnknownError
	}

	// 3. 创建管理员
	admin := &model.Admin{
		Account:    account,
		Password:   string(hashedPassword),
		Name:       name,
		PointName:  pointName,
		Permission: permission,
		Campus:     campus,
	}

	err = adminRepo.Create(ctx, admin)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("创建管理员失败")
		return comm.CodeUnknownError
	}

	// 4. 返回结果
	a.Response = RegisterAdminApiResponse{
		ID: admin.ID,
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (a *RegisterAdminApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&a.Request.Body)
	if err != nil {
		return err
	}
	return err
}

func registerAdmin(ctx *gin.Context) {
	api := &RegisterAdminApi{}
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
