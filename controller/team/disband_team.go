package team

import (
	"gorm.io/gorm"
	"strconv"
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
	teamID := strconv.Itoa(int(team.ID))
	teamSubmitted, _ := global.Rdb.SIsMember(global.Rctx, "teams", teamID).Result()
	if teamSubmitted {
		utility.ResponseError(context, "该队伍已提交，无法解散")
		return
	}

	// 查找团队所有用户
	captain, members := model.GetPersonsInTeam(int(team.ID))

	// 事务
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 删除团队记录
		if err := tx.Delete(&team).Error; err != nil {
			return err
		}

		// 还原所有队员的权限和所属团队ID
		captain.Status = 0
		captain.TeamId = -1

		if err := model.TxUpdatePerson(tx, &captain); err != nil {
			return err
		}
		for _, member := range members {
			member.Status = 0
			member.TeamId = -1
			if err := model.TxUpdatePerson(tx, &member); err != nil {
				return err
			}
		}

		// 返回 nil 提交事务
		return nil
	})
	if err != nil {
		utility.ResponseError(context, "系统异常，请重试")
		return
	}

	utility.SendMessageToMembers(team.Name+"已经被解散", captain, members)

	utility.ResponseSuccess(context, nil)
}
