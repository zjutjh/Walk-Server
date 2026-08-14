package register

import (
	"errors"
	"reflect"
	"runtime"

	"github.com/zjutjh/WeJH-SDK/oauth"
	"github.com/zjutjh/WeJH-SDK/oauth/oauthException"

	peopleCache "app/dao/cache/people"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func RegisterStudentHandler() gin.HandlerFunc {
	api := RegisterStudentApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterStudent).Pointer()).Name()] = api
	return hfRegisterStudent
}

type RegisterStudentApi struct {
	Info     struct{} `name:"学生报名"`
	Request  RegisterStudentApiRequest
	Response struct{}
}

type RegisterStudentApiRequest struct {
	Body struct {
		Name     string `json:"name" desc:"姓名" binding:"required"`
		StuID    string `json:"stu_id" desc:"学号" binding:"required"`
		Password string `json:"password" desc:"密码" binding:"required"`
		Identity string `json:"identity" desc:"身份证号" binding:"required"`
		Tel      string `json:"tel" desc:"电话" binding:"required"`
		Wechat   string `json:"wechat" desc:"微信号"`
		QQ       string `json:"qq" desc:"QQ号"`
	}
}

func (h *RegisterStudentApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request.Body)
}

func (h *RegisterStudentApi) Run(ctx *gin.Context) kit.Code {
	if code := comm.CheckBizPhase(comm.PhaseRegistration); code != comm.CodeOK {
		return code
	}
	info, code := fetchRegisterOAuthInfo(ctx, h.Request.Body.StuID, h.Request.Body.Password)
	if code != comm.CodeOK {
		return code
	}
	if info.UserTypeDesc == "教师职工" || info.UserTypeDesc == "人才派遣" {
		return comm.CodeNonStudentRegister
	}
	if info.Name != h.Request.Body.Name {
		return comm.CodePeopleInfoWrong
	}
	password, err := comm.Hash(h.Request.Body.Password)
	if err != nil {
		return comm.CodeServerError
	}
	identity, err := comm.EncryptIdentity(h.Request.Body.Identity)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("散列身份证号失败")
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
		Tel:        h.Request.Body.Tel,
		IsViolated: false,
		CreatedOp:  3,
		JoinOp:     5,
		TeamID:     -1,
		Type:       comm.MemberTypeStudent,
		WalkStatus: comm.WalkStatusNotStart,
	})
}

func createRegisterPerson(ctx *gin.Context, person *model.People) kit.Code {
	peopleRepo := repo.NewPeopleRepo()
	byIdentity, err := peopleRepo.FindPeopleByStoredIdentity(ctx, person.Identity)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询身份证报名记录失败")
		return comm.CodeServerError
	}
	if byIdentity != nil {
		return comm.CodeAlreadyRegistered
	}

	byTel, err := peopleRepo.FindPeopleByTel(ctx, person.Tel)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询手机号报名记录失败")
		return comm.CodeServerError
	}
	if byTel != nil {
		return comm.CodeAlreadyRegistered
	}

	if person.StuID != "" {
		byStuID, err := peopleRepo.FindPeopleByStuID(ctx, person.StuID)
		if err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("查询学工号报名记录失败")
			return comm.CodeServerError
		}
		if byStuID != nil {
			return comm.CodeAlreadyRegistered
		}
	}

	if err := peopleRepo.Create(ctx, person); err != nil {
		if isRegisterDuplicateError(err) {
			return comm.CodeAlreadyRegistered
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("创建报名记录失败")
		return comm.CodeServerError
	}
	_ = peopleCache.SetPersonByID(ctx, person)
	return comm.CodeOK
}

type registerOAuthInfo struct {
	Name         string
	Gender       string
	College      string
	UserTypeDesc string
}

func fetchRegisterOAuthInfo(ctx *gin.Context, account, password string) (*registerOAuthInfo, kit.Code) {
	cookie, info, err := oauth.GetUserInfo(account, password)
	var oauthErr *oauthException.Error
	if errors.As(err, &oauthErr) {
		if errors.Is(oauthErr, oauthException.WrongAccount) || errors.Is(oauthErr, oauthException.WrongPassword) {
			return nil, comm.CodeAccountOrPasswordError
		}
		nlog.Pick().WithContext(ctx).WithError(err).Warn("统一身份认证失败")
		return nil, comm.CodeOAuthFailed
	}
	if err != nil || cookie == nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("统一身份认证失败")
		return nil, comm.CodeOAuthFailed
	}
	return &registerOAuthInfo{
		Name:         info.Name,
		Gender:       info.Gender,
		College:      info.College,
		UserTypeDesc: info.UserTypeDesc,
	}, comm.CodeOK
}

func isRegisterDuplicateError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
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
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
