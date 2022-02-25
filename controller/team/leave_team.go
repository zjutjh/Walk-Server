package team

import (
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func LeaveTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	var person model.Person
	initial.DB.Where("open_id = ?", jwtData.OpenID).Take(&person)

	if person.Status == 0 {
		utility.ResponseError(context, "请先加入队伍")
		return
	} else if person.Status == 2 {
		utility.ResponseError(context, "队长只能解散队伍")
		return
	}

	// 检查队伍是否提交
	var team model.Team
	initial.DB.Where("id = ?", person.TeamId).Take(&team)

	// 队伍成员数量减一
	result := initial.DB.Model(&team).Where("submitted = 0").Update("num", team.Num-1)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "该队伍已经提交，无法退出")
		return
	}

	// 恢复队员信息到未加入的状态
	person.Status = 0
	person.TeamId = -1
	initial.DB.Save(&person)

	utility.ResponseSuccess(context, nil)
}