package model

import (
	"errors"
	"walk-server/global"
)

type Team struct {
	ID         uint
	Name       string // 队伍的名字
	Num        uint8  // 团队里的人数
	Password   string // 团队加入的密码
	Slogan     string // 团队标语
	AllowMatch bool   // 是否接收随机匹配
	Captain    string // 队长的 Open ID
	Route      uint8  // 1 是朝晖路线，2 屏峰半程，3 屏峰全程，4 莫干山半程，5 莫干山全程
}

func GetTeamInfo(teamID uint) (*Team, error) {
	team := new(Team)
	result := global.DB.Where("id = ?", teamID).Take(team)

	if result.RowsAffected == 0 {
		return nil, errors.New("no team")
	}
	return team, nil
}

func GetPersonsInTeam(teamID int) (Person, []Person) {
	var persons []Person

	var captain Person
	var members []Person

	global.DB.Where("team_id = ?", teamID).Find(&persons)
	for _, person := range persons {
		if person.Status == 2 { // 队长
			captain = person
		} else {
			members = append(members, person)
		}
	}

	return captain, members
}
