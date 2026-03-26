package comm

import "github.com/zjutjh/mygo/kit"

var CodeOK = kit.NewCode(0, "success")

// 系统错误码
var (
	CodeUnknownError           = kit.NewCode(10000, "未知错误")
	CodeThirdServiceError      = kit.NewCode(10001, "三方服务错误")
	CodeDatabaseError          = kit.NewCode(10002, "数据库错误")
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
)

// 业务错误码 从 30000 开始
var (
	CodeAlreadyRegistered      = kit.NewCode(30001, "该身份信息已报名")
	CodeOAuthFailed            = kit.NewCode(30002, "统一身份验证失败")
	CodeIdentityMismatch       = kit.NewCode(30003, "身份信息不匹配")
	CodeAlreadyInTeam          = kit.NewCode(30004, "已在队伍中")
	CodeTeamFull               = kit.NewCode(30005, "队伍人数已满")
	CodeNotInTeam              = kit.NewCode(30006, "尚未加入队伍")
	CodeNotCaptain             = kit.NewCode(30007, "仅队长可操作")
	CodeNoCreateChance         = kit.NewCode(30008, "创建队伍次数已用完")
	CodeNoJoinChance           = kit.NewCode(30009, "加入队伍次数已用完")
	CodeTeamSubmitted          = kit.NewCode(30010, "队伍已提交，无法操作")
	CodeTeamNameDuplicated     = kit.NewCode(30011, "队伍名称已存在")
	CodePasswordWrong          = kit.NewCode(30012, "密码错误")
	CodeTeamNotEnough          = kit.NewCode(30013, "队伍人数不足")
	CodeCannotJoinSelf         = kit.NewCode(30014, "不能添加自己")
	CodeTypeMismatch           = kit.NewCode(30015, "人员类型不匹配，无法加入")
	CodeTeamAlreadySubmit      = kit.NewCode(30016, "队伍已提交")
	CodeOpenIDEmpty            = kit.NewCode(30017, "OpenID为空")
	CodeWechatCodeMissing      = kit.NewCode(30018, "微信Code缺失")
	CodeAccountOrPasswordError = kit.NewCode(30019, "账号或密码错误")
	CodeAccountExistError      = kit.NewCode(30020, "该账号已存在")
)
