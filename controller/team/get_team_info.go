package team

import (
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func GetTeamInfo(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 获取个人信息
	var person model.Person
	initial.DB.Where("open_id = ?", jwtData.OpenID).Take(&person)

	// 先判断是否加入了团队
	if person.Status == 0 {
		utility.ResponseError(context, "尚未加入团队")
		return
	}

	// 查找团队
	var team model.Team
	initial.DB.Where("id = ?", person.TeamId).Take(&team)

	// 查找团队成员
	var persons []model.Person
	var leader model.Person
	var members []gin.H
	initial.DB.Where("team_id = ?", person.TeamId).Find(&persons)
	for _, person := range persons {
		if person.Status == 2 { // 队长
			leader = person
		} else {
			members = append(members, gin.H{
				"name":    person.Name,
				"gender":  person.Gender,
				"open_id": person.OpenId,
				"campus":  person.Campus,
				"contact": gin.H{
					"qq":     person.Qq,
					"wechat": person.Wechat,
					"tel":    person.Tel,
				},
			})
		}
	}

	// 返回结果
	utility.ResponseSuccess(context, gin.H{
		"id":          person.TeamId,
		"name":        team.Name,
		"route":       team.Route,
		"password":    team.Password,
		"submitted":   team.Submitted,
		"allow_match": team.AllowMatch,
		"slogan":      team.Slogan,
		"leader": gin.H{
			"name":    leader.Name,
			"gender":  leader.Gender,
			"campus":  leader.Campus,
			"open_id": leader.OpenId,
			"contact": gin.H{
				"qq":     leader.Qq,
				"wechat": leader.Wechat,
				"tel":    leader.Tel,
			},
		},
		"member": members,
	})
}