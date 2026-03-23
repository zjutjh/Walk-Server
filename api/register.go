package api

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nedis"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/repo"
)

var registerLockReleaseScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`)

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

	lockValue, locked, err := acquireRegisterLock(ctx, openID)
	if err != nil {
		return comm.LogAndCode(ctx, comm.ErrRegisterLockAcquire, err)
	}
	if !locked {
		return comm.CodeTooFrequently
	}
	defer func() {
		if err = releaseRegisterLock(ctx, openID, lockValue); err != nil {
			comm.LogAppError(ctx, comm.ErrRegisterLockRelease, err)
		}
	}()

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
		if isDuplicateEntryError(err) {
			return comm.CodeAlreadyRegistered
		}
		return comm.LogAndCode(ctx, comm.ErrRegisterCreateFailed, err)
	}

	return comm.CodeOK
}

func hfRegisterStudent(ctx *gin.Context) {
	api := &RegisterStudentApi{}
	err := api.Init(ctx)
	if err != nil {
		reply.Fail(ctx, comm.LogAndCode(ctx, comm.ErrRequestBindInvalid, err))
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

func acquireRegisterLock(ctx *gin.Context, openID string) (string, bool, error) {
	lockKey := fmt.Sprintf("walk:user:register:lock:%s", openID)
	lockValue := fmt.Sprintf("%s:%d", openID, time.Now().UnixNano())

	locked, err := nedis.Pick().SetNX(ctx, lockKey, lockValue, comm.BizConf.GetRegisterLockTTL()).Result()
	if err != nil {
		return "", false, err
	}

	return lockValue, locked, nil
}

func releaseRegisterLock(ctx *gin.Context, openID, lockValue string) error {
	if lockValue == "" {
		return nil
	}

	lockKey := fmt.Sprintf("walk:user:register:lock:%s", openID)
	return registerLockReleaseScript.Run(ctx, nedis.Pick(), []string{lockKey}, lockValue).Err()
}

func isDuplicateEntryError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}

func hfRegisterTeacher(ctx *gin.Context) {
	api := &RegisterTeacherApi{}
	err := api.Init(ctx)
	if err != nil {
		reply.Fail(ctx, comm.LogAndCode(ctx, comm.ErrRequestBindInvalid, err))
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
		reply.Fail(ctx, comm.LogAndCode(ctx, comm.ErrRequestBindInvalid, err))
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
