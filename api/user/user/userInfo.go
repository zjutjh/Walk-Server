package user

import (
	"reflect"
	"runtime"

	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func UserInfoHandler() gin.HandlerFunc {
	api := UserInfoApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUserInfo).Pointer()).Name()] = api
	return hfUserInfo
}

type UserInfoApi struct {
	Info     struct{} `name:"获取用户信息"`
	Request  struct{}
	Response UserInfoApiResponse
}

type UserContact struct {
	QQ     string `json:"qq" desc:"QQ号"`
	Wechat string `json:"wechat" desc:"微信号"`
	Tel    string `json:"tel" desc:"电话"`
}

type UserInfoApiResponse struct {
	ID       int         `json:"id"`
	Name     string      `json:"name" desc:"姓名"`
	StuID    string      `json:"stu_id" desc:"学号"`
	Gender   int8        `json:"gender" desc:"1男2女"`
	Campus   string      `json:"campus" desc:"校区 枚举值:'zh''pf''mgs'" `
	College  string      `json:"college" desc:"学院"`
	Role     string      `json:"role" desc:"队伍中身份 枚举值'unbind''captain''member'"`
	CreateOp uint8       `json:"create_op" desc:"剩余创建团队次数"`
	JoinOp   uint8       `json:"join_op" desc:"剩余加入团队次数"`
	TeamID   int64       `json:"team_id" desc:"团队ID"`
	Type     string      `json:"type" desc:"人员类型 枚举值：'alumnus''student''teacher'"`
	Contact  UserContact `json:"contact" desc:"联系方式"`
}

func (h *UserInfoApi) Init(ctx *gin.Context) error { return nil }

func (h *UserInfoApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentUserPerson(ctx)
	if code != comm.CodeOK {
		return code
	}

	h.Response.ID = int(person.ID)
	h.Response.Name = person.Name
	h.Response.StuID = person.StuID
	h.Response.Gender = person.Gender
	h.Response.Campus = person.Campus
	h.Response.College = person.College
	h.Response.Role = person.Role
	h.Response.CreateOp = person.CreatedOp
	h.Response.JoinOp = person.JoinOp
	h.Response.TeamID = person.TeamID
	h.Response.Type = person.Type
	h.Response.Contact = UserContact{
		QQ:     person.Qq,
		Wechat: person.Wechat,
		Tel:    person.Tel,
	}
	return comm.CodeOK
}

func currentUserPerson(ctx *gin.Context) (*model.People, kit.Code) {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return nil, comm.CodeNotLoggedIn
	}
	if comm.BizConf.AESSecret != "" {
		if decrypted, err := comm.AesDecrypt(openID, comm.BizConf.AESSecret); err == nil && decrypted != "" {
			openID = decrypted
		}
	}

	person, err := repo.NewPeopleRepo().FindPeopleByOpenID(ctx, openID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询当前用户失败")
		return nil, comm.CodeServerError
	}
	if person == nil {
		return nil, comm.CodePeopleNotFound
	}
	return person, comm.CodeOK
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
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
