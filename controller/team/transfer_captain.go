package team

import (
	"walk-server/constant"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

type TransferLeaderData struct {
	MemberOpenID string `json:"member_open_id" binding:"required"`
}

func TransferLeader(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 解析 JSON 数据
	var transferLeaderData TransferLeaderData
	err := context.ShouldBindJSON(&transferLeaderData)
	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}

	// 读取成员信息
	member, err := model.GetPerson(transferLeaderData.MemberOpenID)
	if err != nil {
		utility.ResponseError(context, "The member is not exist")
		return
	}

	// 读取队长信息
	captain, _ := model.GetPerson(jwtData.OpenID)
	if captain.Status != constant.IS_CAPTAIN {
		utility.ResponseError(context, "not leader")
		return
	}

	// 更新成员信息
	member.Status = constant.IS_CAPTAIN
	result := global.DB.Model(member).Where("status = ? and team_id = ?",
		constant.IS_MEMBER, captain.TeamId).Save(*member)
	if result.RowsAffected == 0 {
		// 可能不是同一个队伍了
		utility.ResponseError(context, "not the same team")
		return
	}
	// 更新缓存数据
	if _, found := global.Cache.Get(member.OpenId); found {
		global.Cache.SetDefault(member.OpenId, member)
	}

	// 更新队长信息
	captain.Status = constant.IS_MEMBER
	model.UpdatePerson(captain.OpenId, captain)

	// 更新团队信息
	var team model.Team
	global.DB.Where("id = ?", captain.TeamId).Take(&team)
	team.Captain = member.OpenId
	global.DB.Model(&team).Save(team)

	utility.ResponseSuccess(context, nil)
}
