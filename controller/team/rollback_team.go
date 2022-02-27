package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func RollBackTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	person, _ := model.GetPerson(jwtData.OpenID)
	global.DB.Where("open_id = ?", jwtData.OpenID).Take(&person)

	// 判断用户权限
	if person.Status == 0 {
		utility.ResponseError(context, "请先加入队伍")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "没有修改的权限")
		return
	}

	var team model.Team
	var teamCount model.TeamCount

	global.DB.Where("id = ?", person.TeamId).Take(&team)
	if !team.Submitted {
		utility.ResponseError(context, "该队伍还没有提交")
	}

	// 删除队伍的提交状态
	global.DB.Model(&team).Update("submitted", 0)
	global.DB.Where("day_campus = ?", utility.GetCurrentDate()*10+team.Route).Take(&teamCount)
	global.DB.Model(&teamCount).Update("count", teamCount.Count-1)

	utility.ResponseSuccess(context, nil)
}