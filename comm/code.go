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
	CodeInsufficientParams = kit.NewCode(20009, "参数不足")
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
	CodeTeamNotFound           = kit.NewCode(30021, "队伍不存在")
	CodeUserNoQuota            = kit.NewCode(30022, "该用户没有名额")
	CodeBindCodeError          = kit.NewCode(30023, "签到码绑定失败")
	CodePeopleNotFound         = kit.NewCode(30024, "人员不存在")
	CodeCampusMismatch         = kit.NewCode(30025, "校区错误")
	CodeTeamCheckinClosed      = kit.NewCode(30026, "该队伍已完成，无法进行点位打卡")
	CodePrevPointInvalid       = kit.NewCode(30027, "上一签到点并非路线前序点位")
	CodeWrongRouteAlert        = kit.NewCode(30028, "该团队路线走错，请立即提醒")
	CodeTeamMemberInsufficient = kit.NewCode(30029, "团队人数不足")
	CodeTeamMemberExceeded     = kit.NewCode(30030, "团队人数超过上限")
)
