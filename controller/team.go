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

// JoinTeamData 加入团队时接收的信息类型
type JoinTeamData struct {
	TeamID   int    `json:"team_id"`
	Password string `json:"password"`
}

func CreateTeam(context *gin.Context) {
	// TODO 加入对当天团队数上限的判断
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

	openID := jwtData.OpenID
	identify := jwtData.Identity
	var person model.Person

	if identify != "not-join" { // 现在已经加入了一个团队
		utility.ResponseError(context, "请先退出或解散原来的团队")
	}
	initial.DB.Where("open_id = ?", openID).First(&person)
	if person.CreatedOp == 0 {
		utility.ResponseError(context, "无法创建团队了")
	} else {
		// 再数据库中插入一个团队
		team := model.Team{
			Name:      createTeamData.Name,
			Num:       1,
			Password:  createTeamData.Password,
			Captain:   person.Name,
			Route:     createTeamData.Route,
			Submitted: false,
		}
		initial.DB.Create(&team)

		// 将入团队后对应的状态更新
		person.CreatedOp -= 1
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

func JoinTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	if jwtData.Identity != "not-join" { // 如果在一个团队中
		utility.ResponseError(context, "请退出或解散原来的团队")
		return
	}

	var joinTeamData JoinTeamData
	err := context.ShouldBindJSON(&joinTeamData)
	if err != nil { // 参数发送错误
		utility.ResponseError(context, "参数错误")
		return
	}

	// 从数据库中读取用户信息
	var person model.Person
	initial.DB.Where("open_id = ?", jwtData.OpenID).Find(&person)

	if person.JoinOp == 0 { // 加入次数用完了
		utility.ResponseError(context, "没有加入次数了")
		return
	}

	// 检查密码
	var team model.Team
	result := initial.DB.Where("id = ?", joinTeamData.TeamID).First(&team)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "找不到团队")
		return
	}
	if team.Submitted == true {
		utility.ResponseError(context, "该队伍已提交，无法加入")
		return
	}
	if team.Password != joinTeamData.Password {
		utility.ResponseError(context, "密码错误")
		return
	}

	// 如果人数没有大于团队最大人数
	result = initial.DB.Model(&team).Where("num < 6").Update("num", team.Num+1) // 队伍上限 6 人
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "队伍人数到达上限")
	} else {
		person.Status = 1
		person.JoinOp--
		jwtData.Identity = "member"
		jwtData.TeamID = int(team.ID)
		jwtNewToken, _ := utility.GenerateStandardJwt(jwtData)
		initial.DB.Model(&person).Updates(person) // 将新的用户信息写入数据库
		utility.ResponseSuccess(context, gin.H{
			"jwt": jwtNewToken,
		})
	}
}

func GetTeamInfo(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 先判断是否加入了团队
	if jwtData.Identity == "not-join" {
		utility.ResponseError(context, "尚未加入团队")
		return
	}

	// 查找团队
	var team model.Team
	teamID := jwtData.TeamID // 获取队伍信息
	initial.DB.Where("id = ?", teamID).First(&team)

	// 查找团队成员
	var persons []model.Person
	var leader model.Person
	var members []gin.H
	initial.DB.Where("team_id = ?", teamID).Find(&persons)
	for _, person := range persons {
		if person.Status == 2 { // 队长
			leader = person
		} else {
			members = append(members, gin.H{
				"name":   person.Name,
				"gender": person.Gender,
				"contact": gin.H{
					"qq":     person.Qq,
					"wechat": person.Wechat,
					"tel":    person.Tel,
				},
			})
		}
	}

	// 返回结果
	utility.ResponseSuccess(context, gin.H{
		"id":    teamID,
		"name":  team.Name,
		"route": team.Route,
		"leader": gin.H{
			"name":   leader.Name,
			"gender": leader.Gender,
			"contact": gin.H{
				"qq":     leader.Qq,
				"wechat": leader.Wechat,
				"tel":    leader.Tel,
			},
		},
		"member": members,
	})
}
