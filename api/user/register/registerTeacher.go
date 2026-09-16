package register

import (
	"reflect"
	"runtime"

	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func RegisterTeacherHandler() gin.HandlerFunc {
	api := RegisterTeacherApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterTeacher).Pointer()).Name()] = api
	return hfRegisterTeacher
}

type RegisterTeacherApi struct {
	Info     struct{} `name:"教职工报名" `
	Request  RegisterTeacherApiRequest
	Response struct{}
}

type RegisterTeacherApiRequest struct {
	Body struct {
		Name     string `json:"name" desc:"姓名" binding:"required"`
		Identity string `json:"identity" desc:"身份证号" binding:"required"`
		StuID    string `json:"stu_id" desc:"工号" binding:"required"`
		Password string `json:"password" desc:"统一认证密码" binding:"required"`
		Tel      string `json:"tel" desc:"电话" binding:"required"`
		Wechat   string `json:"wechat" desc:"微信号"`
		QQ       string `json:"qq" desc:"QQ号"`
	}
}

func (h *RegisterTeacherApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request.Body)
}

func (h *RegisterTeacherApi) Run(ctx *gin.Context) kit.Code {
	if code := comm.CheckBizPhase(comm.PhaseRegistration, comm.PhaseSubmission, comm.PhaseAdjustment); code != comm.CodeOK {
		return code
	}
	identity := comm.NormalizeIdentity(h.Request.Body.Identity)
	tel := comm.NormalizePhone(h.Request.Body.Tel)
	if !comm.IsValidIdentity(identity) || !comm.IsValidPhone(tel) {
		return comm.CodeParameterInvalid
	}
	info, code := fetchRegisterOAuthInfo(ctx, h.Request.Body.StuID, h.Request.Body.Password)
	if code != comm.CodeOK {
		return code
	}
	if info.UserTypeDesc != "教师职工" {
		return comm.CodeNonTeacherRegister
	}
	if info.Name != h.Request.Body.Name {
		return comm.CodePeopleInfoWrong
	}
	password, err := comm.Hash(h.Request.Body.Password)
	if err != nil {
		return comm.CodeServerError
	}
	identity, err = comm.EncryptIdentity(identity)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("加密身份证号失败")
		return comm.CodeServerError
	}

	return createRegisterPerson(ctx, &model.People{
		Password:   password,
		Name:       info.Name,
		Gender:     comm.ParseGender(info.Gender),
		StuID:      h.Request.Body.StuID,
		Identity:   identity,
		Role:       comm.RoleUnbind,
		Qq:         h.Request.Body.QQ,
		Wechat:     h.Request.Body.Wechat,
		Tel:        tel,
		IsViolated: false,
		CreatedOp:  3,
		JoinOp:     5,
		TeamID:     -1,
		Type:       comm.MemberTypeTeacher,
		WalkStatus: comm.WalkStatusNotStart,
	})
}

func hfRegisterTeacher(ctx *gin.Context) {
	api := &RegisterTeacherApi{}
	if err := api.Init(ctx); err != nil {
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
