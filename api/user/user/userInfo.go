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

// UserContact 用于修改用户信息接口的联系方式请求体。
type UserContact struct {
	QQ     string `json:"qq" desc:"QQ号"`
	Wechat string `json:"wechat" desc:"微信号"`
	Tel    string `json:"tel" desc:"电话"`
}

type UserInfoApiResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name" desc:"姓名"`
	Gender   int8   `json:"gender" desc:"性别 0未知1男2女"`
	StuID    string `json:"stu_id" desc:"学号或工号"`
	Tel      string `json:"tel" desc:"电话"`
	Wechat   string `json:"wechat" desc:"微信号"`
	QQ       string `json:"qq" desc:"QQ号"`
	Role     string `json:"role" desc:"队伍中身份 枚举值'unbind''captain''member'"`
	CreateOp uint8  `json:"create_op" desc:"剩余创建团队次数"`
	JoinOp   uint8  `json:"join_op" desc:"剩余加入团队次数"`
	TeamID   int64  `json:"team_id" desc:"团队ID"`
	Type     string `json:"type" desc:"人员类型 枚举值：'alumnus''student''teacher'"`
}

func (h *UserInfoApi) Init(ctx *gin.Context) error { return nil }

func (h *UserInfoApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentUserPerson(ctx)
	if code != comm.CodeOK {
		return code
	}

	h.Response.ID = int(person.ID)
	h.Response.Name = person.Name
	h.Response.Gender = person.Gender
	h.Response.StuID = person.StuID
	h.Response.Tel = person.Tel
	h.Response.Wechat = person.Wechat
	h.Response.QQ = person.Qq
	h.Response.Role = person.Role
	h.Response.CreateOp = person.CreatedOp
	h.Response.JoinOp = person.JoinOp
	h.Response.TeamID = person.TeamID
	h.Response.Type = person.Type
	return comm.CodeOK
}

func currentUserPerson(ctx *gin.Context) (*model.People, kit.Code) {
	userID, err := comm.GetUserIDFromCtx(ctx)
	if err != nil || userID <= 0 {
		return nil, comm.CodeNotLoggedIn
	}
	person, err := repo.NewPeopleRepo().FindPeopleByID(ctx, userID)
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
