package team

import (
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func RollBackTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	var person model.Person
	initial.DB.Where("open_id = ?", jwtData.OpenID).Take(&person)

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

	initial.DB.Where("id = ?", person.TeamId).Take(&team)
	if !team.Submitted {
		utility.ResponseError(context, "该队伍还没有提交")
	}

	// 删除队伍的提交状态
	initial.DB.Model(&team).Update("submitted", 0)
	initial.DB.Where("day_campus = ?", utility.GetCurrentDate()*10+team.Route).Take(&teamCount)
	initial.DB.Model(&teamCount).Update("count", teamCount.Count-1)

	utility.ResponseSuccess(context, nil)
}