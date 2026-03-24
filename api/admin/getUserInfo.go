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
	"app/dao/repo"
)

func GetUserInfoByIDHandler() gin.HandlerFunc {
	api := GetUserInfoByIDApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(getUserInfoByID).Pointer()).Name()] = api
	return getUserInfoByID
}

type GetUserInfoByIDApi struct {
	Info     struct{} `name:"获取人员信息"`
	Request  GetUserInfoByIDApiRequest
	Response GetUserInfoByIDApiResponse
}

type GetUserInfoByIDApiRequest struct {
	Query struct {
		UserID int `form:"user_id" binding:"required"`
	}
}

type GetUserInfoByIDApiResponse struct {
	Name string `json:"name" desc:"用户姓名"`
}

// Run Api业务逻辑执行点
func (g *GetUserInfoByIDApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()
	peopleRepo := repo.NewPeopleRepo()

	user, err := peopleRepo.FindPeopleByID(ctx, int64(g.Request.Query.UserID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询人员信息失败")
		return comm.CodeUnknownError
	}
	if user == nil {
		return comm.CodeDataNotFound
	}
	if user.TeamID <= 0 {
		return comm.CodeUserNoQuota
	}

	team, err := teamRepo.FindTeamByID(ctx, user.TeamID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍信息失败")
		return comm.CodeUnknownError
	}
	if team == nil || team.Submit != 1 {
		return comm.CodeUserNoQuota
	}

	g.Response = GetUserInfoByIDApiResponse{
		Name: user.Name,
	}

	return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (g *GetUserInfoByIDApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

func getUserInfoByID(ctx *gin.Context) {
	api := &GetUserInfoByIDApi{}
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
