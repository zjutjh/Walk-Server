package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// UpdateTeamData 更新团队信息的数据类型
type UpdateTeamData struct {
	Name       string `json:"name" binding:"required"`
	Route      uint8  `json:"route" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Slogan     string `json:"slogan" binding:"required"`
	AllowMatch bool   `json:"allow_match"`
}

func UpdateTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	var person model.Person
	global.DB.Where("open_id = ?", jwtData.OpenID).Take(&person)

	// 判断用户权限
	if person.Status == 0 {
		utility.ResponseError(context, "请先加入队伍")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "没有修改的权限")
		return
	}

	// 解析 post 数据
	var updateTeamData UpdateTeamData
	err := context.ShouldBindJSON(&updateTeamData)
	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}

	// 更新团队信息
	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)
	if team.Submitted {
		utility.ResponseError(context, "该队伍已经提交，无法修改")
		return
	}
	team.Name = updateTeamData.Name
	team.Route = updateTeamData.Route
	team.Password = updateTeamData.Password
	team.AllowMatch = updateTeamData.AllowMatch
	team.Slogan = updateTeamData.Slogan
	global.DB.Save(&team)
	utility.ResponseSuccess(context, nil)
}
