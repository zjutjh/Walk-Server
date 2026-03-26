package comm

const (
	WalkStatusNotStart  = "notStart"
	WalkStatusPending   = "pending"
	WalkStatusAbandoned = "abandoned"
	WalkStatusViolated  = "violated"
	WalkStatusCompleted = "completed"
)

const (
	RoleUnbind  = "unbind"
	RoleCaptain = "captain"
	RoleMember  = "member"
)

const (
	CodeChekin = "checkin"
	CodeTeam   = "team"
)

const (
	AdminPermissionSuper    = "super"
	AdminPermissionManager  = "manager"
	AdminPermissionInternal = "internal"
	AdminPermissionExternal = "external"
)

const (
	TeamStatusNotStart  = "notStart"
	TeamStatusCompleted = "completed"
	TeamStatusWithDrawn = "withDrawn"
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

type WalkStatus string

const (
	WalkStatusNotStarted WalkStatus = "not_started"
	WalkStatusReady      WalkStatus = "ready"
	WalkStatusInProgress WalkStatus = "in_progress"
	WalkStatusQuit       WalkStatus = "quit"
	WalkStatusWithdrawn  WalkStatus = "withdrawn"
	WalkStatusViolation  WalkStatus = "violation"
	WalkStatusFinished   WalkStatus = "finished"
)

type TeamStatus string

const (
	TeamStatusNotStarted TeamStatus = "not_started"
	TeamStatusInProgress TeamStatus = "in_progress"
	TeamStatusFinished   TeamStatus = "finished"
	TeamStatusWithdrawn  TeamStatus = "withdrawn"
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
		return string(WalkStatusNotStarted)
	case 2:
		return string(WalkStatusReady)
	case 3:
		return string(WalkStatusInProgress)
	case 4:
		return string(WalkStatusQuit)
	case 5:
		return string(WalkStatusWithdrawn)
	case 6:
		return string(WalkStatusViolation)
	case 7:
		return string(WalkStatusFinished)
	default:
		return ""
	}
}

func FormatTeamStatus(value uint8) string {
	switch value {
	case 1:
		return string(TeamStatusNotStarted)
	case 2:
		return string(TeamStatusInProgress)
	case 3:
		return string(TeamStatusFinished)
	case 4:
		return string(TeamStatusWithdrawn)
	default:
		return ""
	}
}
