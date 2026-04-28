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
	CodeCheckin = CodeChekin
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
	TeamStatusWithdrawn  = "withdrawn"
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

func IsValidWalkStatus(status string) bool {
	switch status {
	case WalkStatusNotStart,
		WalkStatusPending,
		WalkStatusAbandoned,
		WalkStatusInProgress,
		WalkStatusWithdrawn,
		WalkStatusViolated,
		WalkStatusCompleted:
		return true
	default:
		return false
	}
}

func IsValidRole(role string) bool {
	switch role {
	case RoleUnbind, RoleCaptain, RoleMember:
		return true
	default:
		return false
	}
}

func IsValidCodeType(codeType string) bool {
	switch codeType {
	case CodeChekin, CodeTeam:
		return true
	default:
		return false
	}
}

func IsValidAdminPermission(permission string) bool {
	switch permission {
	case AdminPermissionSuper,
		AdminPermissionManager,
		AdminPermissionInternal,
		AdminPermissionExternal:
		return true
	default:
		return false
	}
}

func IsValidTeamStatus(status string) bool {
	switch status {
	case TeamStatusNotStart, TeamStatusInProgress, TeamStatusCompleted, TeamStatusWithdrawn:
		return true
	default:
		return false
	}
}

func IsValidGender(gender string) bool {
	switch gender {
	case GenderMale, GenderFemale:
		return true
	default:
		return false
	}
}

func IsValidCampus(campus string) bool {
	switch campus {
	case CampusChaohui, CampusPingfeng, CampusMoganshan:
		return true
	default:
		return false
	}
}

func IsValidMemberType(memberType string) bool {
	switch memberType {
	case MemberTypeStudent, MemberTypeTeacher, MemberTypeAlumnus:
		return true
	default:
		return false
	}
}

func IsValidTeamRole(role string) bool {
	switch role {
	case TeamRoleNone, TeamRoleMember, TeamRoleCaptain:
		return true
	default:
		return false
	}
}

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
		return TeamStatusWithdrawn
	default:
		return ""
	}
}
