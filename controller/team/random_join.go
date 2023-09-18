package team

import (
	"gorm.io/gorm"
	"strconv"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

type RandomJoinData struct {
	ID int `json:"id" binding:"required"`
}

// RandomJoin 随机组队
func RandomJoin(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 读取用户信息
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status != 0 { // 如果在一个团队中
		utility.ResponseError(context, "请退出或解散原来的团队")
		return
	}

	if person.JoinOp == 0 { // 加入次数用完了
		utility.ResponseError(context, "没有加入次数了")
		return
	}

	// 解析 JSON 数据
	var randomJoinData RandomJoinData
	err := context.ShouldBindJSON(&randomJoinData)
	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}

	// 加入队伍
	var team model.Team
	global.DB.Where("id = ?", randomJoinData.ID).Take(&team)
	teamID := strconv.Itoa(int(team.ID))
	teamSubmitted, _ := global.Rdb.SIsMember(global.Rctx, "teams", teamID).Result()
	if teamSubmitted {
		utility.ResponseError(context, "队伍刚刚提交了")
		return
	}

	if team.Num >= 6 || !team.AllowMatch {
		utility.ResponseError(context, "队伍刚刚满人了或者关闭了随机组队")
		return
	}

	// 获取这个团队原来的队长和队员
	captain, members := model.GetPersonsInTeam(int(team.ID))

	err = global.DB.Transaction(func(tx *gorm.DB) error {
		// 队伍成员数量加一
		if err := tx.Model(&team).Update("num", team.Num+1).Error; err != nil {
			return err
		}

		// 更新加入成员的信息
		person.Status = 1
		person.JoinOp--
		person.TeamId = int(team.ID)
		if err := model.TxUpdatePerson(tx, person); err != nil {
			return err
		}

		return nil
	})

	// 加入成功以后发送消息给所有的用户
	utility.SendMessageToTeam(person.Name+"通过随机组队加入了队伍", captain, members)

	utility.ResponseSuccess(context, nil)
}
