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

func RegisterStudentHandler() gin.HandlerFunc {
	api := RegisterStudentApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterStudent).Pointer()).Name()] = api
	return hfRegisterStudent
}

func RegisterTeacherHandler() gin.HandlerFunc {
	api := RegisterTeacherApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterTeacher).Pointer()).Name()] = api
	return hfRegisterTeacher
}

func RegisterAlumnusHandler() gin.HandlerFunc {
	api := RegisterAlumnusApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterAlumnus).Pointer()).Name()] = api
	return hfRegisterAlumnus
}

type RegisterCommonRequest struct {
	Name     string `json:"name" binding:"required"`
	Gender   string `json:"gender" binding:"required" desc:"字符串枚举: male|female"`
	Campus   string `json:"campus" binding:"required" desc:"字符串枚举: chaohui|pingfeng|moganshan"`
	StuID    string `json:"stu_id"`
	Identity string `json:"identity" binding:"required"`
	QQ       string `json:"qq"`
	Wechat   string `json:"wechat"`
	College  string `json:"college" binding:"required"`
	Tel      string `json:"tel" binding:"required"`
}

type RegisterStudentApi struct {
	Info     struct{} `name:"学生报名" desc:"学生报名接口"`
	Request  RegisterCommonRequest
	Response struct{}
}

type RegisterTeacherApi struct {
	Info     struct{} `name:"教职工报名" desc:"教职工报名接口"`
	Request  RegisterCommonRequest
	Response struct{}
}

type RegisterAlumnusApi struct {
	Info     struct{} `name:"校友报名" desc:"校友报名接口"`
	Request  RegisterCommonRequest
	Response struct{}
}

func (h *RegisterStudentApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *RegisterTeacherApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *RegisterAlumnusApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *RegisterStudentApi) Run(ctx *gin.Context) kit.Code {
	return doRegister(ctx, h.Request, 1)
}

func (h *RegisterTeacherApi) Run(ctx *gin.Context) kit.Code {
	return doRegister(ctx, h.Request, 2)
}

func (h *RegisterAlumnusApi) Run(ctx *gin.Context) kit.Code {
	return doRegister(ctx, h.Request, 3)
}

func doRegister(ctx *gin.Context, req RegisterCommonRequest, personType uint8) kit.Code {
	gender, ok := parseGender(req.Gender)
	if !ok {
		return comm.CodeParameterInvalid
	}

	campus, ok := parseCampus(req.Campus)
	if !ok {
		return comm.CodeParameterInvalid
	}

	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	existing, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if existing != nil {
		return comm.CodeAlreadyRegistered
	}

	byIdentity, err := peopleRepo.FindByIdentity(ctx, req.Identity)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if byIdentity != nil {
		return comm.CodeAlreadyRegistered
	}

	if personType == 1 && req.StuID != "" {
		byStuID, err := peopleRepo.FindByStuID(ctx, req.StuID)
		if err != nil {
			return comm.CodeDatabaseError
		}
		if byStuID != nil {
			return comm.CodeAlreadyRegistered
		}
	}

	err = peopleRepo.Create(ctx, &repo.Person{
		OpenID:     openID,
		Name:       req.Name,
		Gender:     gender,
		StuID:      req.StuID,
		Campus:     campus,
		Identity:   req.Identity,
		Role:       0,
		QQ:         req.QQ,
		Wechat:     req.Wechat,
		College:    req.College,
		Tel:        req.Tel,
		CreatedOp:  3,
		JoinOp:     5,
		TeamID:     -1,
		Type:       personType,
		WalkStatus: 1,
	})
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("创建报名记录失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func hfRegisterStudent(ctx *gin.Context) {
	api := &RegisterStudentApi{}
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
			reply.Success(ctx, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}

func hfRegisterAlumnus(ctx *gin.Context) {
	api := &RegisterAlumnusApi{}
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
