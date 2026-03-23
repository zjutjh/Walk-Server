package comm

// 人员活动状态枚举
const (
	WalkStatusNotStart   = "notStart"   // 未开始
	WalkStatusPending    = "pending"    // 待出发
	WalkStatusAbandoned  = "abandoned"  // 已放弃
	WalkStatusInProgress = "inProgress" // 进行中
	WalkStatusWithdrawn  = "withdrawn"  // 已下撤
	WalkStatusViolated   = "violated"   // 已违规
	WalkStatusCompleted  = "completed"  // 已完成
)

//role枚举
const (
	RoleUnbind= "unbind"  // 未绑定
	RoleCaptain = "captain" // 队长
	RoleMember  = "member"  // 队员
)

//codeType枚举
const(
	CodeChekin = "checkin" //签到码
	CodeTeam = "team" //团队码
)

//admin_permission枚举
const(
	AdminPermissionSuper = "super" //超级管理员
	AdminPermissionManager = "manager" //管理员
	AdminPermissionInternal = "internal" //内部人员
	AdminPermissionExternal = "external" //外部人员
)

//TeamStatus枚举
const(
	TeamStatusNotStart = "notStart" //未开始
	TeamStatusInProgress = "inProgress" //进行中
	TeamStatusCompleted = "completed" //已完成
	TeamStatusWithDrawn = "withDrawn" //已下撤
)