package team

import (
	"strconv"
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

	// 判断用户权限
	if person.Status == 0 {
		utility.ResponseError(context, "请先加入队伍")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "没有修改的权限")
		return
	}

	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)
	teamID := strconv.Itoa(int(team.ID))
	teamSubmitted, _ := global.Rdb.SIsMember(global.Rctx, "teams", teamID).Result()
	if !teamSubmitted {
		utility.ResponseError(context, "队伍未提交")
		return
	}

	// 删除队伍的提交状态
	global.Rdb.SRem(global.Rctx, "teams", teamID)
	dailyRoute := utility.GetCurrentDate()*10 + team.Route
	dailyRouteKey := strconv.Itoa(int(dailyRoute))
	global.Rdb.Incr(global.Rctx, dailyRouteKey)

	utility.ResponseSuccess(context, nil)
}
