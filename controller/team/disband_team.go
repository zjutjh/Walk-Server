package team

import (
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func DisbandTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	var person model.Person
	initial.DB.Where("open_id = ?", jwtData.OpenID).Take(&person)

	if person.Status == 0 {
		utility.ResponseError(context, "请先创建一个队伍")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "队员无法解散队伍")
		return
	}

	// 查找团队
	var team model.Team
	initial.DB.Where("id = ?", person.TeamId).Take(&team)

	if team.Submitted {
		utility.ResponseError(context, "该队伍已提交，无法解散")
		return
	}

	// 查找团队所有用户
	var persons []model.Person
	initial.DB.Where("team_id = ?", person.TeamId).Find(&persons)

	// 删除团队记录
	initial.DB.Delete(&team)

	// 还原所有队员的权限和所属团队ID
	for _, person := range persons {
		person.Status = 0
		person.TeamId = -1
		initial.DB.Save(&person)
	}

	utility.ResponseSuccess(context, nil)
}