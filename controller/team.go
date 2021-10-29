package controller

import (
	"github.com/gin-gonic/gin"
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"
)

// CreateTeamData 接收创建团队信息的数据类型
type CreateTeamData struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Route    uint8  `json:"route" binding:"required"`
}

func CreateTeam(context *gin.Context) {
	// 获取 open ID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 获取 post json 数据
	var createTeamData CreateTeamData
	err := context.ShouldBindJSON(&createTeamData)
	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}

	openID := jwtData.OpenID
	identify := jwtData.Identity
	var person model.Person

	if identify != "not-join" { // 现在已经加入了一个团队
		utility.ResponseError(context, "请先退出原来的团队")
	}
	initial.DB.Where("open_id = ?", openID).First(&person)
	if person.CreatedOp == 0 {
		utility.ResponseError(context, "无法创建团队了")
	} else {
		// 再数据库中插入一个团队
		team := model.Team{
			Name:      createTeamData.Name,
			Password:  createTeamData.Password,
			Captain:   person.Name,
			Route:     createTeamData.Route,
			Submitted: false,
		}
		initial.DB.Create(&team)

		// 将入团队后对应的状态更新
		person.Status = 1
		jwtData.TeamID = int(team.ID)
		jwtData.Identity = "leader"
		jwtNewToken, _ := utility.GenerateStandardJwt(jwtData)

		initial.DB.Model(&person).Updates(person) // 将新的用户信息写入数据库

		// 返回新的 team_id 和 jwt 数据
		utility.ResponseSuccess(context, gin.H{
			"team_id": team.ID,
			"jwt":     jwtNewToken,
		})
	}
}
