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

func UserInfoHandler() gin.HandlerFunc {
	api := UserInfoApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUserInfo).Pointer()).Name()] = api
	return hfUserInfo
}

func UserModifyHandler() gin.HandlerFunc {
	api := UserModifyApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUserModify).Pointer()).Name()] = api
	return hfUserModify
}

type UserInfoApi struct {
	Info     struct{} `name:"用户信息" desc:"获取当前登录用户信息"`
	Request  struct{}
	Response UserInfoApiResponse
}

type UserInfoApiResponse struct {
	Person *repo.Person `json:"person"`
	Team   *repo.Team   `json:"team"`
}

type UserModifyApi struct {
	Info     struct{} `name:"修改用户信息" desc:"修改当前登录用户可编辑信息"`
	Request  UserModifyApiRequest
	Response struct{}
}

type UserModifyApiRequest struct {
	QQ     string `json:"qq"`
	Wechat string `json:"wechat"`
	Tel    string `json:"tel"`
}

func (h *UserInfoApi) Init(ctx *gin.Context) (err error) {
	return err
}

func (h *UserModifyApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *UserInfoApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil {
		return comm.CodeDataNotFound
	}

	h.Response.Person = person
	if person.TeamID > 0 {
		team, err := teamRepo.FindByID(ctx, person.TeamID)
		if err != nil {
			return comm.CodeDatabaseError
		}
		h.Response.Team = team
	}

	return comm.CodeOK
}

func (h *UserModifyApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	err := repo.NewPeopleRepo().UpdateByOpenID(ctx, openID, map[string]any{
		"qq":     h.Request.QQ,
		"wechat": h.Request.Wechat,
		"tel":    h.Request.Tel,
	})
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("更新用户信息失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func hfUserInfo(ctx *gin.Context) {
	api := &UserInfoApi{}
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
			reply.Success(ctx, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
