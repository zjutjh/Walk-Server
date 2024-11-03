package team

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"
)

// 获取请求参数
type request struct {
	OpenID string `json:"open_id"`
}

func ChangeCaptain(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	person, _ := model.GetPerson(jwtData.OpenID)

	// 判断用户权限
	if person.Status == 0 {
		utility.ResponseError(context, "请先加入队伍")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "没有修改的权限")
		return
	}

	// 获取队伍
	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)

	// 判断是否队长
	if team.Captain != person.OpenId {
		utility.ResponseError(context, "不是队长")
		return
	}

	// 判断队伍是否提交
	teamID := strconv.Itoa(int(team.ID))
	teamSubmitted, _ := global.Rdb.SIsMember(global.Rctx, "teams", teamID).Result()
	if teamSubmitted {
		utility.ResponseError(context, "该队伍已经提交，无法修改")
		return
	}

	// 获取请求参数
	var data request
	err := context.ShouldBindJSON(&data)
	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}

	// 查找新队长
	newCaptain, err := model.GetPerson(data.OpenID)
	if err != nil {
		utility.ResponseError(context, "未找到用户")
		return
	}

	// 判断新队长是否在队伍中
	if newCaptain.TeamId != person.TeamId {
		utility.ResponseError(context, "不在队伍中")
		return
	}

	if person.Type != 1 && newCaptain.Type == 1 {
		utility.ResponseError(context, "无法将队长移交给学生")
		return

	}

	// 更换队长
	team.Captain = newCaptain.OpenId
	global.DB.Save(&team)

	person.Status = 1
	model.UpdatePerson(jwtData.OpenID, person)

	newCaptain.Status = 2
	model.UpdatePerson(data.OpenID, newCaptain)

	utility.ResponseSuccess(context, nil)
}
