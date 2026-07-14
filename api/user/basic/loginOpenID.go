package basic

import (
	"net/url"
	"reflect"
	"runtime"
	"strings"

	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func WechatLoginByOpenIDHandler() gin.HandlerFunc {
	api := WechatLoginByOpenIDApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfWechatLoginByOpenID).Pointer()).Name()] = api
	return hfWechatLoginByOpenID
}

type WechatLoginByOpenIDApi struct {
	Info     struct{} `name:"OpenID登录" desc:"通过OpenID换取系统JWT"`
	Request  WechatLoginByOpenIDApiRequest
	Response WechatLoginByOpenIDApiResponse
}

type WechatLoginByOpenIDApiRequest struct {
	Query struct {
		OpenID string `form:"open_id" desc:"微信OpenID" binding:"required"`
	}
}

type WechatLoginByOpenIDApiResponse struct {
	JWT  string                   `json:"jwt" desc:"系统JWT"`
	User *WechatLoginByOpenIDUser `json:"user" desc:"用户信息"`
}

type WechatLoginByOpenIDUser struct {
	ID         int64  `json:"id" desc:"人员ID"`
	OpenID     string `json:"open_id" desc:"微信OpenID"`
	Name       string `json:"name" desc:"姓名"`
	StuID      string `json:"stu_id" desc:"学号或工号"`
	Gender     int8   `json:"gender" desc:"性别 1男2女"`
	Campus     string `json:"campus" desc:"校区 枚举值:'zh''pf''mgs'"`
	College    string `json:"college" desc:"学院"`
	Role       string `json:"role" desc:"队伍中身份 枚举值:'unbind''member''captain'"`
	CreateOp   uint8  `json:"create_op" desc:"剩余创建团队次数"`
	JoinOp     uint8  `json:"join_op" desc:"剩余加入团队次数"`
	TeamID     int64  `json:"team_id" desc:"团队ID"`
	Type       string `json:"type" desc:"人员类型 枚举值:'student''teacher''alumnus'"`
	WalkStatus string `json:"walk_status" desc:"人员活动状态 枚举值:'not_start''pending''abandoned''in_progress''withdrawn''violated''completed'"`
	QQ         string `json:"qq" desc:"QQ号"`
	Wechat     string `json:"wechat" desc:"微信号"`
	Tel        string `json:"tel" desc:"电话"`
}

func (h *WechatLoginByOpenIDApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindQuery(&h.Request.Query)
}

func (h *WechatLoginByOpenIDApi) Run(ctx *gin.Context) kit.Code {
	openID, err := url.QueryUnescape(h.Request.Query.OpenID)
	if err != nil {
		return comm.CodeParameterInvalid
	}
	openID = strings.ReplaceAll(openID, " ", "+")
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	jwt, err := comm.GenerateToken(openID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("生成用户JWT失败")
		return comm.CodeServerError
	}
	h.Response.JWT = jwt

	person, err := repo.NewPeopleRepo().FindPeopleByOpenID(ctx, openID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询OpenID用户失败")
		return comm.CodeServerError
	}
	h.Response.User = buildWechatLoginByOpenIDUser(person)
	return comm.CodeOK
}

func buildWechatLoginByOpenIDUser(person *model.People) *WechatLoginByOpenIDUser {
	if person == nil {
		return nil
	}
	return &WechatLoginByOpenIDUser{
		ID:         person.ID,
		OpenID:     person.OpenID,
		Name:       person.Name,
		StuID:      person.StuID,
		Gender:     person.Gender,
		Campus:     person.Campus,
		College:    person.College,
		Role:       person.Role,
		CreateOp:   person.CreatedOp,
		JoinOp:     person.JoinOp,
		TeamID:     person.TeamID,
		Type:       person.Type,
		WalkStatus: person.WalkStatus,
		QQ:         person.Qq,
		Wechat:     person.Wechat,
		Tel:        person.Tel,
	}
}

func hfWechatLoginByOpenID(ctx *gin.Context) {
	api := &WechatLoginByOpenIDApi{}
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
