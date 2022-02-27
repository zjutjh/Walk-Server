package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// CreateTeamData 接收创建团队信息的数据类型
type CreateTeamData struct {
	Name       string `json:"name" binding:"required"`
	Route      uint8  `json:"route" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Slogan     string `json:"slogan" binding:"required"`
	AllowMatch bool   `json:"allow_match"`
}

func CreateTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 获取 post json 数据
	var createTeamData CreateTeamData
	err := context.ShouldBindJSON(&createTeamData)
	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}

	// 查询用户信息
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status != 0 { // 现在已经加入了一个团队
		utility.ResponseError(context, "请先退出或解散原来的团队")
		return
	}

	if person.CreatedOp == 0 {
		utility.ResponseError(context, "无法创建团队了")
		return
	} else {
		// 再数据库中插入一个团队
		team := model.Team{
			Name:       createTeamData.Name,
			Num:        1,
			AllowMatch: createTeamData.AllowMatch,
			Password:   createTeamData.Password,
			Captain:    person.OpenId,
			Route:      createTeamData.Route,
			Slogan:     createTeamData.Slogan,
			Submitted:  false,
		}
		global.DB.Create(&team)

		// 将入团队后对应的状态更新
		person.CreatedOp -= 1
		person.Status = 2
		person.TeamId = int(team.ID)

		model.UpdatePerson(jwtData.OpenID, person)

		// 返回新的 team_id 和 jwt 数据
		utility.ResponseSuccess(context, gin.H{
			"team_id": team.ID,
		})
	}
}
