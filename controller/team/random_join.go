package team

import (
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
	global.DB.Where("id = ?", randomJoinData.ID).Take(&team)                                                          // 取出这个队伍
	captain, members := model.GetPersonsInTeam(int(team.ID))                                                          // 获取这个团队原来的队长和队员
	result := global.DB.Model(&team).Where("num < 6 AND allow_match = 1 AND submitted = 0").Update("num", team.Num+1) // 更新队伍人数
	if result.RowsAffected != 0 {                                                                                     // 更新成功
		person.Status = 1
		person.JoinOp--
		person.TeamId = int(team.ID)
		model.UpdatePerson(jwtData.OpenID, person) // 将新的用户信息写入数据库
		utility.ResponseSuccess(context, nil)

		// 加入成功以后发送消息给所有的用户
		utility.SendMessageToTeam(person.Name+"通过随机组队加入了队伍", captain, members)
		return
	}

	utility.ResponseError(context, "队伍刚刚满人了或者关闭了随机组队")
}
