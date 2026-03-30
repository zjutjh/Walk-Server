package api

import (
	"app/dao/model"
	"errors"
)

var (
	errTeamNameDuplicated = errors.New("team name duplicated")
	errTeamJoinConflict   = errors.New("team join conflict")
	errTeamLeaveConflict  = errors.New("team leave conflict")
)

type TeamInfoApiResponse struct {
	Team    *TeamInfoTeamView    `json:"team"`
	Members []TeamInfoMemberView `json:"members"`
}

type TeamInfoTeamView struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Num           uint8  `json:"num"`
	Slogan        string `json:"slogan"`
	AllowMatch    bool   `json:"allow_match"`
	Captain       string `json:"captain"`
	RouteName     string `json:"route_name"`
	PrevPointName string `json:"prev_point_name"`
	Submit        bool   `json:"submit"`
	Status        string `json:"status" desc:"字符串枚举: notStart|inProgress|completed|withDrawn"`
	IsLost        bool   `json:"is_lost"`
}

type TeamInfoMemberView struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Gender     string `json:"gender" desc:"字符串枚举: male|female"`
	StuID      string `json:"stu_id"`
	Campus     string `json:"campus" desc:"字符串枚举: zh|pf|mgs"`
	Identity   string `json:"identity"`
	TeamRole   string `json:"team_role" desc:"字符串枚举: unbind|member|captain"`
	QQ         string `json:"qq"`
	Wechat     string `json:"wechat"`
	College    string `json:"college"`
	Tel        string `json:"tel"`
	TeamID     int64  `json:"team_id"`
	MemberType string `json:"member_type" desc:"字符串枚举: student|teacher|alumnus"`
	WalkStatus string `json:"walk_status" desc:"字符串枚举: notStart|pending|abandoned|inProgress|withdrawn|violated|completed"`
}

func toTeamInfoTeamView(team *model.Team) *TeamInfoTeamView {
	if team == nil {
		return nil
	}
	return &TeamInfoTeamView{
		ID:            team.ID,
		Name:          team.Name,
		Num:           uint8(team.Num),
		Slogan:        team.Slogan,
		AllowMatch:    team.AllowMatch != 0,
		Captain:       team.Captain,
		RouteName:     team.RouteName,
		PrevPointName: team.PrevPointName,
		Submit:        team.Submit != 0,
		Status:        team.Status,
		IsLost:        team.IsLost != 0,
	}
}

func toTeamInfoMemberViews(members []model.People) []TeamInfoMemberView {
	result := make([]TeamInfoMemberView, 0, len(members))
	for _, member := range members {
		result = append(result, TeamInfoMemberView{
			ID:         member.ID,
			Name:       member.Name,
			Gender:     formatGender(member.Gender),
			StuID:      member.StuID,
			Campus:     member.Campus,
			Identity:   member.Identity,
			TeamRole:   member.Role,
			QQ:         member.Qq,
			Wechat:     member.Wechat,
			College:    member.College,
			Tel:        member.Tel,
			TeamID:     member.TeamID,
			MemberType: member.Type,
			WalkStatus: member.WalkStatus,
		})
	}
	return result
}

func formatGender(value int8) string {
	if value == 1 {
		return "male"
	}
	if value == 2 {
		return "female"
	}
	return ""
}

func boolToInt8(value bool) int8 {
	if value {
		return 1
	}
	return 0
}
