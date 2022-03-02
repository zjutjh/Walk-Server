package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func LeaveTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status == 0 {
		utility.ResponseError(context, "请先加入队伍")
		return
	} else if person.Status == 2 {
		utility.ResponseError(context, "队长只能解散队伍")
		return
	}

	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)

	// 队伍成员数量减一
	result := global.DB.Model(&team).Where("submitted = 0").Update("num", team.Num-1)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "该队伍已经提交，无法退出")
		return
	}

	// 恢复队员信息到未加入的状态
	person.Status = 0
	person.TeamId = -1
	model.UpdatePerson(jwtData.OpenID, person)

	captain, members := model.GetPersonsInTeam(int(team.ID)) // 获取这个人退出了以后团队中的所有成员
	utility.SendMessageToTeam(person.Name + "已经离开了队伍", captain, members)

	utility.ResponseSuccess(context, nil)
}