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
	CodeCheckin = "checkin"
	CodeTeam    = "team"
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
