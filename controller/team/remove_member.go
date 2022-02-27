package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func RemoveMember(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	var person model.Person
	global.DB.Where("open_id = ?", jwtData.OpenID).Take(&person)

	if person.Status == 0 {
		utility.ResponseError(context, "请先加入团队")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "只有队长可以移除队员")
		return
	}

	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)
	if team.Submitted {
		utility.ResponseError(context, "该队伍已经提交, 无法移除队员")
		return
	}

	// 读取 Get 参数
	memberRemovedOpenID := context.Query("openid")

	var personRemoved model.Person
	result := global.DB.Where("open_id = ?", memberRemovedOpenID).Take(&personRemoved)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "没有这个用户")
		return
	} else if personRemoved.TeamId != person.TeamId {
		utility.ResponseError(context, "不能移除别的队伍的人")
		return
	}

	// 队伍数量减少
	team.Num--
	global.DB.Save(&team)

	// 更新被踢出的人的状态
	personRemoved.Status = 0
	personRemoved.TeamId = -1
	global.DB.Save(&personRemoved)
}