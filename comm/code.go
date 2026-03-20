package comm

import "github.com/zjutjh/mygo/kit"

var CodeOK = kit.NewCode(0, "success")

// 系统错误码
var (
	CodeUnknownError           = kit.NewCode(10000, "未知错误")
	CodeThirdServiceError      = kit.NewCode(10001, "三方服务错误")
	CodeDatabaseError          = kit.NewCode(10002, "数据库错误")
	CodeRedisError             = kit.NewCode(10003, "Redis错误")
	CodeMiddlewareServiceError = kit.NewCode(10004, "中间件服务错误")
)

// 业务通用错误码
var (
	CodeNotLoggedIn        = kit.NewCode(20000, "用户未登录")
	CodeLoginExpired       = kit.NewCode(20001, "登录过期，请重新登录")
	CodePermissionDenied   = kit.NewCode(20002, "用户无权限")
	CodeParameterInvalid   = kit.NewCode(20003, "参数错误")
	CodeDataParseError     = kit.NewCode(20004, "数据解析异常")
	CodeDataNotFound       = kit.NewCode(20005, "数据不存在")
	CodeDataConflict       = kit.NewCode(20006, "数据冲突")
	CodeServiceMaintenance = kit.NewCode(20007, "系统维护中")
	CodeTooFrequently      = kit.NewCode(20008, "操作过于频繁/未获得锁")
)

// 业务错误码 从 30000 开始
var (
	CodeAccountOrPasswordError = kit.NewCode(30000, "账号或密码错误")
	CodeAccountExistError      = kit.NewCode(30001, "该账号已存在")
	CodeTeamNotFound           = kit.NewCode(30002, "队伍不存在")
	CodeUserNoQuota            = kit.NewCode(30003, "该用户没有名额")
	CodeBindCodeError          = kit.NewCode(30004, "签到码绑定失败")
)
