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
	Info     struct{} `name:"教职工报名" desc:"教职工报名接口"`
	Request  RegisterTeacherApiRequest
	Response struct{}
}

type RegisterTeacherApiRequest struct {
	Body struct {
		Identity string          `json:"identity" desc:"身份证号" binding:"required"`
		StuID    string          `json:"stu_id" desc:"工号" binding:"required"`
		Password string          `json:"password" desc:"密码" binding:"required"`
		Campus   string          `json:"campus" desc:"校区枚举值"`
		Contact  RegisterContact `json:"contact" desc:"联系方式" binding:"required"`
	}
}

func (h *RegisterTeacherApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request.Body)
}

func (h *RegisterTeacherApi) Run(ctx *gin.Context) kit.Code {
	campus := ""
	if h.Request.Body.Campus != "" {
		parsedCampus, ok := comm.ParseCampus(h.Request.Body.Campus)
		if !ok {
			return comm.CodeParameterInvalid
		}
		campus = parsedCampus
	}

	info, code := fetchRegisterOAuthInfo(ctx, h.Request.Body.StuID, h.Request.Body.Password)
	if code != comm.CodeOK {
		return code
	}
	if info.UserTypeDesc != "教师职工" && info.UserTypeDesc != "人才派遣" {
		return comm.CodeParameterInvalid
	}

	return createRegisterPerson(ctx, &model.People{
		Name:       info.Name,
		Gender:     comm.ParseGender(info.Gender),
		StuID:      h.Request.Body.StuID,
		Campus:     campus,
		Identity:   h.Request.Body.Identity,
		Role:       comm.RoleUnbind,
		Qq:         h.Request.Body.Contact.QQ,
		Wechat:     h.Request.Body.Contact.Wechat,
		College:    info.College,
		Tel:        h.Request.Body.Contact.Tel,
		CreatedOp:  2,
		JoinOp:     5,
		TeamID:     -1,
		Type:       comm.MemberTypeTeacher,
		WalkStatus: comm.WalkStatusNotStart,
	})
}

func hfRegisterTeacher(ctx *gin.Context) {
	api := &RegisterTeacherApi{}
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
