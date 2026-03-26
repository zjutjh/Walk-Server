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

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type Campus string

const (
	CampusChaohui   Campus = "chaohui"
	CampusPingfeng  Campus = "pingfeng"
	CampusMoganshan Campus = "moganshan"
)

type MemberType string

const (
	MemberTypeStudent MemberType = "student"
	MemberTypeTeacher MemberType = "teacher"
	MemberTypeAlumnus MemberType = "alumnus"
)

type TeamRole string

const (
	TeamRoleNone    TeamRole = "none"
	TeamRoleMember  TeamRole = "member"
	TeamRoleCaptain TeamRole = "captain"
)

func ParseGender(raw string) (int8, bool) {
	switch Gender(raw) {
	case GenderMale:
		return 1, true
	case GenderFemale:
		return 2, true
	default:
		return 0, false
	}
}

func ParseCampus(raw string) (uint8, bool) {
	switch Campus(raw) {
	case CampusChaohui:
		return 1, true
	case CampusPingfeng:
		return 2, true
	case CampusMoganshan:
		return 3, true
	default:
		return 0, false
	}
}

func FormatGender(value int8) string {
	switch value {
	case 1:
		return string(GenderMale)
	case 2:
		return string(GenderFemale)
	default:
		return ""
	}
}

func FormatCampus(value uint8) string {
	switch value {
	case 1:
		return string(CampusChaohui)
	case 2:
		return string(CampusPingfeng)
	case 3:
		return string(CampusMoganshan)
	default:
		return ""
	}
}

func FormatMemberType(value uint8) string {
	switch value {
	case 1:
		return string(MemberTypeStudent)
	case 2:
		return string(MemberTypeTeacher)
	case 3:
		return string(MemberTypeAlumnus)
	default:
		return ""
	}
}

func FormatTeamRole(value uint8) string {
	switch value {
	case 0:
		return string(TeamRoleNone)
	case 1:
		return string(TeamRoleMember)
	case 2:
		return string(TeamRoleCaptain)
	default:
		return ""
	}
}

func FormatWalkStatus(value uint8) string {
	switch value {
	case 1:
		return WalkStatusNotStart
	case 2:
		return WalkStatusPending
	case 3:
		return WalkStatusInProgress
	case 4:
		return WalkStatusAbandoned
	case 5:
		return WalkStatusWithdrawn
	case 6:
		return WalkStatusViolated
	case 7:
		return WalkStatusCompleted
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
