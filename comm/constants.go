package comm

const (
	// Person Types
	PersonTypeStudent = 1 // 学生
	PersonTypeTeacher = 2 // 教师
	PersonTypeAlumnus = 3 // 校友

	// Person Status (Team related)
	PersonStatusNone    = 0 // 未加入队伍
	PersonStatusMember  = 1 // 队员
	PersonStatusCaptain = 2 // 队长
)

const (
	// Team Status
	TeamStatusNormal = 0 // 正常
)

const (
	// Error Messages
	MsgStudentAlreadyRegistered = "该学号已报名"
	MsgTeacherAlreadyRegistered = "该工号已报名"
	MsgDataNotFound             = "数据不存在"
	MsgDataConflict             = "数据冲突"
	MsgPermissionDenied         = "权限不足"
)

const (
	MaxTeamMember = 6
)
