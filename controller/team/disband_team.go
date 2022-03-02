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
	captain, members := model.GetPersonsInTeam(int(team.ID))

	// 删除团队记录
	global.DB.Delete(&team)

	// 还原所有队员的权限和所属团队ID
	captain.Status = 0
	captain.TeamId = -1
	model.UpdatePerson(captain.OpenId, &captain)
	for _, member := range members {
		member.Status = 0
		member.TeamId = -1
		model.UpdatePerson(member.OpenId, &member)
	}

	utility.SendMessageToMembers(team.Name + "已经被解散", captain, members)

	utility.ResponseSuccess(context, nil)
}