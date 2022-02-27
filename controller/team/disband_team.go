package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func DisbandTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status == 0 {
		utility.ResponseError(context, "请先创建一个队伍")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "队员无法解散队伍")
		return
	}

	// 查找团队
	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)

	if team.Submitted {
		utility.ResponseError(context, "该队伍已提交，无法解散")
		return
	}

	// 查找团队所有用户
	var persons []model.Person
	global.DB.Where("team_id = ?", person.TeamId).Find(&persons)

	// 删除团队记录
	global.DB.Delete(&team)

	// 还原所有队员的权限和所属团队ID
	for _, person := range persons {
		person.Status = 0
		person.TeamId = -1
		model.UpdatePerson(person.OpenId, &person)
	}

	utility.ResponseSuccess(context, nil)
}