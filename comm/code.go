package comm

import "github.com/zjutjh/mygo/kit"

var CodeOK = kit.NewCode(200, "success")

// 系统错误码
var (
	CodeUnknownError           = kit.NewCode(20000, "未知错误")
	CodeDatabaseError          = kit.NewCode(20001, "数据库错误")
	CodeMiddlewareServiceError = kit.NewCode(20002, "中间件服务错误")
)

// 业务通用错误码
var (
	CodeNotLoggedIn        = kit.NewCode(20003, "用户未登录")
	CodePermissionDenied   = kit.NewCode(20004, "用户无权限")
	CodeParameterInvalid   = kit.NewCode(20005, "参数错误")
	CodeDataNotFound       = kit.NewCode(20006, "数据不存在")
	CodeDataConflict       = kit.NewCode(20007, "数据冲突")
	CodeTooFrequently      = kit.NewCode(20008, "操作过于频繁/未获得锁")
	CodeInsufficientParams = kit.NewCode(20009, "参数不足")
)

// 业务错误码 从 20010 开始
var (
	CodeAlreadyRegistered       = kit.NewCode(20010, "该身份信息已报名")
	CodeOAuthFailed             = kit.NewCode(20011, "统一身份验证失败")
	CodeAlreadyInTeam           = kit.NewCode(20012, "已在队伍中")
	CodeTeamFull                = kit.NewCode(20013, "队伍人数已满")
	CodeNotInTeam               = kit.NewCode(20014, "尚未加入队伍")
	CodeNotCaptain              = kit.NewCode(20015, "仅队长可操作")
	CodeNoCreateChance          = kit.NewCode(20016, "创建队伍次数已用完")
	CodeNoJoinChance            = kit.NewCode(20017, "加入队伍次数已用完")
	CodeTeamSubmitted           = kit.NewCode(20018, "队伍已提交，无法操作")
	CodeTeamNameDuplicated      = kit.NewCode(20019, "队伍名称已存在")
	CodePasswordWrong           = kit.NewCode(20020, "密码错误")
	CodeTeamNotEnough           = kit.NewCode(20021, "队伍人数不足")
	CodeOpenIDEmpty             = kit.NewCode(20022, "OpenID为空")
	CodeWechatCodeMissing       = kit.NewCode(20023, "微信Code缺失")
	CodeAccountOrPasswordError  = kit.NewCode(20024, "账号或密码错误")
	CodeTeamNotFound            = kit.NewCode(20025, "队伍不存在")
	CodeUserNoQuota             = kit.NewCode(20026, "该用户没有名额")
	CodeBindCodeError           = kit.NewCode(20027, "签到码绑定失败")
	CodePeopleNotFound          = kit.NewCode(20028, "人员不存在")
	CodeCampusMismatch          = kit.NewCode(20029, "校区错误")
	CodeTeamCheckinClosed       = kit.NewCode(20030, "该队伍已完成，无法进行点位打卡")
	CodePrevPointInvalid        = kit.NewCode(20031, "上一签到点并非路线前序点位")
	CodeWrongRouteAlert         = kit.NewCode(20032, "该团队路线走错，请立即提醒")
	CodeDuplicateCheckin        = kit.NewCode(20033, "该点位已打卡，请勿重复打卡")
	CodeAdminLoginTooFrequently = kit.NewCode(20034, "登录失败次数过多，请稍后再试")
)
