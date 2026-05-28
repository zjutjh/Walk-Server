package comm

import "github.com/zjutjh/mygo/kit"

var CodeOK = kit.NewCode(200, "success")

// 系统错误码
var (
	CodeUnknownError           = kit.NewCode(200101, "未知错误")
	CodeDatabaseError          = kit.NewCode(200102, "数据库错误")
	CodeMiddlewareServiceError = kit.NewCode(200103, "中间件服务错误")
)

// 业务通用错误码
var (
	CodeNotLoggedIn        = kit.NewCode(200201, "未登录或登录过期")
	CodePermissionDenied   = kit.NewCode(200202, "用户无权限")
	CodeParameterInvalid   = kit.NewCode(200203, "参数错误")
	CodeDataNotFound       = kit.NewCode(200204, "数据不存在")
	CodeDataConflict       = kit.NewCode(200205, "数据冲突")
	CodeTooFrequently      = kit.NewCode(200206, "操作过于频繁/未获得锁")
	CodeInsufficientParams = kit.NewCode(200207, "参数不足")
)

// 业务错误码 从 20010 开始
var (
	CodeAlreadyRegistered       = kit.NewCode(200301, "该身份信息已报名")
	CodeOAuthFailed             = kit.NewCode(200302, "统一身份验证失败")
	CodeAlreadyInTeam           = kit.NewCode(200303, "已在团队中")
	CodeTeamFull                = kit.NewCode(200304, "团队人数已满")
	CodeNotInTeam               = kit.NewCode(200305, "尚未加入团队")
	CodeNotCaptain              = kit.NewCode(200306, "仅队长可操作")
	CodeNoCreateChance          = kit.NewCode(200307, "创建团队次数已用完")
	CodeNoJoinChance            = kit.NewCode(200308, "加入团队次数已用完")
	CodeTeamSubmitted           = kit.NewCode(200309, "团队已提交，无法操作")
	CodeTeamNameDuplicated      = kit.NewCode(200310, "团队名称已存在")
	CodePasswordWrong           = kit.NewCode(200311, "密码错误")
	CodeTeamNotEnough           = kit.NewCode(200312, "团队人数不足")
	CodeOpenIDEmpty             = kit.NewCode(200313, "OpenID为空")
	CodeWechatCodeMissing       = kit.NewCode(200314, "微信Code缺失")
	CodeAccountOrPasswordError  = kit.NewCode(200315, "账号或密码错误")
	CodeTeamNotFound            = kit.NewCode(200316, "团队不存在")
	CodeUserNoQuota             = kit.NewCode(200317, "该用户没有名额")
	CodeBindCodeError           = kit.NewCode(200318, "签到码绑定失败")
	CodePeopleNotFound          = kit.NewCode(200319, "人员不存在")
	CodeCampusMismatch          = kit.NewCode(200320, "校区错误")
	CodeAdminLoginTooFrequently = kit.NewCode(200322, "登录失败次数过多，请稍后再试")
	CodeTeamDirectionInvalid    = kit.NewCode(200323, "团队行进方向错误")
)
