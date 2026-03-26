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
	repo "app/dao/repo/admin"
)

func GetUserInfoByScanHandler() gin.HandlerFunc {
	api := GetUserInfoByScanApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(getUserInfoByScan).Pointer()).Name()] = api
	return getUserInfoByScan
}

type GetUserInfoByScanApi struct {
	Info     struct{} `name:"扫码获取人员信息"`
	Request  GetUserInfoByScanApiRequest
	Response GetUserInfoByScanApiResponse
}

type GetUserInfoByScanApiRequest struct {
	Query struct {
		Content string `form:"content" binding:"required"`
	}
}

type GetUserInfoByScanApiResponse struct {
	Name   string `json:"name" desc:"用户姓名"`
	UserID int    `json:"user_id" desc:"用户编号"`
}

// Run Api业务逻辑执行点
func (g *GetUserInfoByScanApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()
	peopleRepo := repo.NewPeopleRepo()

	user, err := peopleRepo.FindPeopleByOpenID(ctx, g.Request.Query.Content)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("根据OpenID查询人员信息失败")
		return comm.CodeDatabaseError
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
		return comm.CodeDatabaseError
	}
	if team == nil || team.Submit != 1 {
		return comm.CodeUserNoQuota
	}

	g.Response = GetUserInfoByScanApiResponse{
		Name:   user.Name,
		UserID: int(user.ID),
	}

	return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (g *GetUserInfoByScanApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&g.Request.Query)
	if err != nil {
		return err
	}
	return err
}

func getUserInfoByScan(ctx *gin.Context) {
	api := &GetUserInfoByScanApi{}
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
