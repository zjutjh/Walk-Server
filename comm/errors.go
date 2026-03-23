package comm

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
)

type AppError struct {
	err   error
	code  kit.Code
	level string
}

func NewAppError(code kit.Code, level, message string) *AppError {
	return &AppError{
		err:   errors.New(message),
		code:  code,
		level: level,
	}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.err.Error()
}

func (e *AppError) Code() kit.Code {
	if e == nil {
		return CodeUnknownError
	}
	return e.code
}

var (
	ErrRequestBindInvalid   = NewAppError(CodeParameterInvalid, "warn", "参数绑定校验错误")
	ErrRegisterLockAcquire  = NewAppError(CodeRedisError, "error", "获取报名锁失败")
	ErrRegisterLockRelease  = NewAppError(CodeRedisError, "warn", "释放报名锁失败")
	ErrRegisterCreateFailed = NewAppError(CodeDatabaseError, "error", "创建报名记录失败")
)

func LogAppError(ctx *gin.Context, appErr *AppError, err error) {
	if appErr == nil {
		return
	}

	logger := nlog.Pick().WithContext(ctx)
	if err != nil {
		logger = logger.WithError(err)
	}

	switch appErr.level {
	case "warn":
		logger.Warn(appErr.Error())
	default:
		logger.Error(appErr.Error())
	}
}

func LogAndCode(ctx *gin.Context, appErr *AppError, err error) kit.Code {
	LogAppError(ctx, appErr, err)
	return appErr.Code()
}
