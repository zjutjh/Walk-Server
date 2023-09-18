package team

import (
	"gorm.io/gorm"
	"strconv"
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
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status == 0 {
		utility.ResponseError(context, "请先加入团队")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "只有队长可以移除队员")
		return
	}

	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)
	teamID := strconv.Itoa(int(team.ID))
	teamSubmitted, _ := global.Rdb.SIsMember(global.Rctx, "teams", teamID).Result()
	if teamSubmitted {
		utility.ResponseError(context, "该队伍已经提交, 无法移除队员")
		return
	}

	// 读取 Get 参数
	memberRemovedOpenID := context.Query("openid")

	personRemoved, err := model.GetPerson(memberRemovedOpenID)
	if err != nil {
		utility.ResponseError(context, "没有这个用户")
		return
	} else if personRemoved.TeamId != person.TeamId {
		utility.ResponseError(context, "不能移除别的队伍的人")
		return
	}

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		// 队伍成员数量减一
		if err := tx.Model(&team).Update("num", team.Num-1).Error; err != nil {
			return err
		}

		// 更新被踢出的人的状态
		personRemoved.Status = 0
		personRemoved.TeamId = -1
		if err := model.TxUpdatePerson(tx, personRemoved); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		utility.ResponseError(context, "服务异常，请重试")
		return
	}

	// 通知被踢出的人
	utility.SendMessage("你被团队"+team.Name+"踢出", nil, personRemoved)

	utility.ResponseSuccess(context, nil)
}
