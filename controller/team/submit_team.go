package team

import (
	"fmt"
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func SubmitTeam(context *gin.Context) {
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
	if team.Submitted {
		utility.ResponseError(context, "该队伍已经提交过了")
	}

	// 开始提交
	tx := initial.DB.Begin() // 开始事务
	tx.Where("day_campus = ?", utility.GetCurrentDate()*10+team.Route).Take(&teamCount)
	key := fmt.Sprintf("teamUpperLimit.%v.%v", team.Route, utility.GetCurrentDate())
	result := tx.Model(&teamCount).Where("count < ?", initial.Config.GetInt(key)).Update("count", teamCount.Count+1)
	if result.RowsAffected == 0 { // 队伍数量到达上限
		utility.ResponseError(context, "队伍数量已经到达上限，无法提交")
		tx.Commit()
	} else { // 团队提交状态更新
		team.Submitted = true
		result := tx.Model(&team).Where("num >= 4").Update("submitted", 1)
		if result.RowsAffected == 0 {
			utility.ResponseError(context, "队伍人数不足 4 人")
			tx.Rollback() // 人数不够回滚 teamCount
		} else {
			utility.ResponseSuccess(context, nil)
			tx.Commit()
		}
	}
}