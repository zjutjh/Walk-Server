package comm

// 人员活动状态枚举
const (
	WalkStatusNotStart   = "notStart"
	WalkStatusPending    = "pending"
	WalkStatusAbandoned  = "abandoned"
	WalkStatusInProgress = "inProgress"
	WalkStatusWithdrawn  = "withdrawn"
	WalkStatusViolated   = "violated"
	WalkStatusCompleted  = "completed"
)

// role枚举
const (
	RoleUnbind  = "unbind"
	RoleCaptain = "captain"
	RoleMember  = "member"
)

// codeType枚举
const (
	CodeChekin = "checkin"
	CodeTeam   = "team"
)

// admin_permission枚举
const (
	AdminPermissionSuper    = "super"
	AdminPermissionManager  = "manager"
	AdminPermissionInternal = "internal"
	AdminPermissionExternal = "external"
)

// TeamStatus枚举
const (
	TeamStatusNotStart   = "notStart"
	TeamStatusInProgress = "inProgress"
	TeamStatusCompleted  = "completed"
	TeamStatusWithDrawn  = "withDrawn"
)

const (
	GenderMale   = "male"
	GenderFemale = "female"
)

const (
	CampusChaohui   = "zh"
	CampusPingfeng  = "pf"
	CampusMoganshan = "mgs"
)

const (
	MemberTypeStudent = "student"
	MemberTypeTeacher = "teacher"
	MemberTypeAlumnus = "alumnus"
)

const (
	TeamRoleNone    = "none"
	TeamRoleMember  = "member"
	TeamRoleCaptain = "captain"
)

func ParseGender(raw string) (int8, bool) {
	switch raw {
	case GenderMale:
		return 1, true
	case GenderFemale:
		return 2, true
	default:
		return 0, false
	}
}

func ParseCampus(raw string) (string, bool) {
	switch raw {
	case CampusChaohui:
		return CampusChaohui, true
	case CampusPingfeng:
		return CampusPingfeng, true
	case CampusMoganshan:
		return CampusMoganshan, true
	default:
		return "", false
	}
}

func FormatGender(value int8) string {
	switch value {
	case 1:
		return GenderMale
	case 2:
		return GenderFemale
	default:
		return ""
	}
}

func FormatTeamStatus(value uint8) string {
	switch value {
	case 1:
		return TeamStatusNotStart
	case 2:
		return TeamStatusInProgress
	case 3:
		return TeamStatusCompleted
	case 4:
		return TeamStatusWithDrawn
	default:
		return ""
	}
}
