package comm

// 人员活动状态枚举
const (
	WalkStatusNotStart   = "not_start"
	WalkStatusPending    = "pending"
	WalkStatusAbandoned  = "abandoned"
	WalkStatusInProgress = "in_progress"
	WalkStatusWithdrawn  = "withdrawn"
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
	TeamStatusNotStart   = "not_start"
	TeamStatusInProgress = "in_progress"
	TeamStatusCompleted  = "completed"
	TeamStatusWithdrawn  = "withdrawn"
)

const (
	CampusChaohui   = "zh"
	CampusPingfeng  = "pf"
	CampusMoganshan = "mgs"
)

const (
	GenderMale   int8 = 1
	GenderFemale int8 = 2
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
		WalkStatusCompleted:
		return true
	default:
		return false
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

func ParseGender(raw string) int8 {
	if raw == "male" {
		return GenderMale
	}
	return GenderFemale
}
