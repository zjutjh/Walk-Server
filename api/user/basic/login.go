package basic

import (
	"reflect"
	"runtime"
	"strings"

	"app/comm"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
)

func LoginHandler() gin.HandlerFunc {
	api := LoginApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfLogin).Pointer()).Name()] = api
	return hfLogin
}

type LoginApi struct {
	Info     struct{} `name:"用户登录" desc:"使用手机号或学工号和报名密码登录"`
	Request  LoginApiRequest
	Response LoginApiResponse
}

type LoginApiRequest struct {
	Body struct {
		AccountType comm.LoginAccountType `json:"account_type" desc:"账号类型：tel、stu_id" binding:"required,oneof=tel stu_id"`
		Account     string                `json:"account" desc:"手机号或学工号" binding:"required"`
		Password    string                `json:"password" desc:"密码" binding:"required"`
	}
}

type LoginApiResponse struct {
	JWT  string     `json:"jwt" desc:"系统JWT"`
	User *LoginUser `json:"user" desc:"用户信息"`
}

type LoginUser struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	StuID      string `json:"stu_id"`
	Role       string `json:"role"`
	CreateOp   uint8  `json:"create_op"`
	JoinOp     uint8  `json:"join_op"`
	TeamID     int64  `json:"team_id"`
	Type       string `json:"type"`
	WalkStatus string `json:"walk_status"`
	QQ         string `json:"qq"`
	Wechat     string `json:"wechat"`
	Tel        string `json:"tel"`
}

func (h *LoginApi) Init(ctx *gin.Context) error { return ctx.ShouldBindJSON(&h.Request.Body) }

func (h *LoginApi) Run(ctx *gin.Context) kit.Code {
	account := strings.TrimSpace(h.Request.Body.Account)
	peopleRepo := repo.NewPeopleRepo()
	var person *model.People
	var err error
	switch h.Request.Body.AccountType {
	case comm.LoginAccountTel:
		account = comm.NormalizePhone(account)
		if !comm.IsValidPhone(account) {
			return comm.CodeParameterInvalid
		}
		person, err = peopleRepo.FindPeopleByTel(ctx, account)
	case comm.LoginAccountStuID:
		person, err = peopleRepo.FindPeopleByStuID(ctx, account)
	default:
		return comm.CodeParameterInvalid
	}
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询登录用户失败")
		return comm.CodeServerError
	}
	if person == nil || person.Password == "" || !comm.Verify(person.Password, h.Request.Body.Password) {
		return comm.CodeAccountOrPasswordError
	}
	token, err := comm.GenerateToken(person.ID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("生成用户JWT失败")
		return comm.CodeServerError
	}
	h.Response.JWT = token
	h.Response.User = buildLoginUser(person)
	return comm.CodeOK
}

func buildLoginUser(p *model.People) *LoginUser {
	return &LoginUser{ID: p.ID, Name: p.Name, StuID: p.StuID,
		Role: p.Role, CreateOp: p.CreatedOp, JoinOp: p.JoinOp,
		TeamID: p.TeamID, Type: p.Type, WalkStatus: p.WalkStatus, QQ: p.Qq, Wechat: p.Wechat, Tel: p.Tel}
}

func hfLogin(ctx *gin.Context) {
	api := &LoginApi{}
	if err := api.Init(ctx); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	if code := api.Run(ctx); code == comm.CodeOK {
		reply.Reply(ctx, comm.CodeOK, api.Response)
	} else {
		reply.Fail(ctx, code)
	}
}
