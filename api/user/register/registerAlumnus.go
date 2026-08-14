package register

import (
	"errors"
	"reflect"
	"runtime"

	"app/comm"
	peopleCache "app/dao/cache/people"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"gorm.io/gorm"
)

func RegisterAlumnusHandler() gin.HandlerFunc {
	api := RegisterAlumnusApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterAlumnus).Pointer()).Name()] = api
	return hfRegisterAlumnus
}

type RegisterAlumnusApi struct {
	Info     struct{} `name:"校友注册"`
	Request  RegisterAlumnusApiRequest
	Response struct{}
}

type RegisterAlumnusApiRequest struct {
	Body struct {
		Name     string `json:"name" desc:"姓名" binding:"required"`
		Identity string `json:"identity" desc:"身份证号" binding:"required"`
		Tel      string `json:"tel" desc:"电话" binding:"required"`
		Password string `json:"password" desc:"登录密码" binding:"required"`
	}
}

func (h *RegisterAlumnusApi) Init(ctx *gin.Context) error { return ctx.ShouldBindJSON(&h.Request.Body) }

func (h *RegisterAlumnusApi) Run(ctx *gin.Context) kit.Code {
	peopleRepo := repo.NewPeopleRepo()
	person, err := peopleRepo.FindPeopleByIdentity(ctx, h.Request.Body.Identity)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询校友身份信息失败")
		return comm.CodeServerError
	}
	if person == nil || person.Type != comm.MemberTypeAlumnus {
		return comm.CodePeopleNotFound
	}
	// 预导入校友只按姓名、身份证号和电话号码核验。
	if person.Name != h.Request.Body.Name || person.Tel != h.Request.Body.Tel {
		return comm.CodePeopleInfoWrong
	}
	if person.Password != "" {
		return comm.CodeAlreadyRegistered
	}
	hashed, err := comm.Hash(h.Request.Body.Password)
	if err != nil {
		return comm.CodeServerError
	}
	if err := peopleRepo.CompleteAlumnusRegistration(ctx, person.ID, hashed); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || isRegisterDuplicateError(err) {
			return comm.CodeAlreadyRegistered
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("完成校友注册失败")
		return comm.CodeServerError
	}
	_ = peopleCache.DelPersonByID(ctx, person.ID)
	return comm.CodeOK
}

func hfRegisterAlumnus(ctx *gin.Context) {
	api := &RegisterAlumnusApi{}
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
