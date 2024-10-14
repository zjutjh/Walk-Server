package team

import (
	"gorm.io/gorm"
	"strconv"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// JoinTeamData 加入团队时接收的信息类型
type JoinTeamData struct {
	TeamID   int    `json:"team_id"`
	Password string `json:"password"`
}

func JoinTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	var joinTeamData JoinTeamData
	err := context.ShouldBindJSON(&joinTeamData)
	if err != nil { // 参数发送错误
		utility.ResponseError(context, "参数错误")
		return
	}

	// 从数据库中读取用户信息
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status != 0 { // 如果在一个团队中
		utility.ResponseError(context, "请退出或解散原来的团队")
		return
	}

	if person.JoinOp == 0 { // 加入次数用完了
		utility.ResponseError(context, "没有加入次数了")
		return
	}

	// 检查密码
	var team model.Team
	result := global.DB.Where("id = ?", joinTeamData.TeamID).Take(&team)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "找不到团队")
		return
	} else if team.Password != joinTeamData.Password {
		utility.ResponseError(context, "密码错误")
		return
	}

	teamID := strconv.Itoa(int(team.ID))
	teamSubmitted, _ := global.Rdb.SIsMember(global.Rctx, "teams", teamID).Result()
	if teamSubmitted {
		utility.ResponseError(context, "该队伍已提交，无法加入")
		return
	}

	// 队伍上限 6 人
	if team.Num >= 6 {
		utility.ResponseError(context, "队伍人数到达上限")
		return
	}

	// 获取这个团队原来的队长和队员
	captain, members := model.GetPersonsInTeam(int(team.ID))

	if captain.Type == 1 && person.Type == 2 {
		utility.ResponseError(context, "您是教师，无法加入学生队伍")
		return
	}

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
	utility.SendMessageToTeam(person.Name+"加入了团队", captain, members)

	utility.ResponseSuccess(context, nil)
}
