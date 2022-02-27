package team

import (
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
	var person model.Person
	global.DB.Where("open_id = ?", jwtData.OpenID).Find(&person)

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
	}
	if team.Submitted {
		utility.ResponseError(context, "该队伍已提交，无法加入")
		return
	}
	if team.Password != joinTeamData.Password {
		utility.ResponseError(context, "密码错误")
		return
	}

	// 如果人数没有大于团队最大人数
	result = global.DB.Model(&team).Where("num < 6").Update("num", team.Num+1) // 队伍上限 6 人
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "队伍人数到达上限")
	} else {
		person.Status = 1
		person.JoinOp--
		person.TeamId = int(team.ID)
		global.DB.Model(&person).Updates(person) // 将新的用户信息写入数据库
		utility.ResponseSuccess(context, nil)
	}
}