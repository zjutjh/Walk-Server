package register

import (
	"reflect"
	"runtime"

	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func RegisterAlumnusHandler() gin.HandlerFunc {
	api := RegisterAlumnusApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfRegisterAlumnus).Pointer()).Name()] = api
	return hfRegisterAlumnus
}

type RegisterAlumnusApi struct {
	Info     struct{} `name:"校友登录" desc:"校友登录绑定当前OpenID"`
	Request  RegisterAlumnusApiRequest
	Response struct{}
}

type RegisterAlumnusApiRequest struct {
	Body struct {
		Name     string `json:"name" desc:"姓名" binding:"required"`
		Identity string `json:"identity" desc:"身份证号" binding:"required"`
		Tel      string `json:"tel" desc:"电话" binding:"required"`
	}
}

func (h *RegisterAlumnusApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request.Body)
}

func (h *RegisterAlumnusApi) Run(ctx *gin.Context) kit.Code {
	if !comm.IsValidIdentity(h.Request.Body.Identity) {
		return comm.CodeParameterInvalid
	}

	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeNotLoggedIn
	}
	if comm.BizConf.AESSecret != "" {
		if decrypted, err := comm.AesDecrypt(openID, comm.BizConf.AESSecret); err == nil && decrypted != "" {
			openID = decrypted
		}
	}

	peopleRepo := repo.NewPeopleRepo()
	person, err := peopleRepo.FindPeopleByIdentity(ctx, h.Request.Body.Identity)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询校友身份信息失败")
		return comm.CodeServerError
	}
	if person == nil {
		return comm.CodePeopleNotFound
	}
	if person.Tel != h.Request.Body.Tel && person.Name != h.Request.Body.Name {
		return comm.CodeParameterInvalid
	}

	bound, err := peopleRepo.FindPeopleByOpenID(ctx, openID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询OpenID绑定信息失败")
		return comm.CodeServerError
	}
	if bound != nil && bound.ID != person.ID {
		return comm.CodeAlreadyRegistered
	}

	if err := peopleRepo.BindAlumnusOpenID(ctx, person, openID); err != nil {
		if isRegisterDuplicateError(err) {
			return comm.CodeAlreadyRegistered
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("绑定校友OpenID失败")
		return comm.CodeServerError
	}

	return comm.CodeOK
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
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
