package team

import (
	"strconv"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func GetTeamInfo(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 获取个人信息
	person, _ := model.GetPerson(jwtData.OpenID)

	// 先判断是否加入了团队
	if person.Status == 0 {
		utility.ResponseError(context, "尚未加入团队")
		return
	}

	// 查找团队
	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)

	// 查找团队成员
	captain, members := model.GetPersonsInTeam(person.TeamId)

	// 获取团队成员响应信息
	captainData := gin.H{
		"name":    captain.Name,
		"gender":  captain.Gender,
		"campus":  captain.Campus,
		"open_id": captain.OpenId,
		"type":    captain.Type,
		"contact": gin.H{
			"qq":     captain.Qq,
			"wechat": captain.Wechat,
			"tel":    captain.Tel,
		},
		"walk_status": captain.WalkStatus,
	}
	var memberData []gin.H
	for _, member := range members {
		memberData = append(memberData, gin.H{
			"name":    member.Name,
			"gender":  member.Gender,
			"open_id": member.OpenId,
			"campus":  member.Campus,
			"type":    member.Type,
			"contact": gin.H{
				"qq":     member.Qq,
				"wechat": member.Wechat,
				"tel":    member.Tel,
			},
			"walk_status": member.WalkStatus,
		})
	}

	teamID := strconv.Itoa(int(team.ID))
	teamSubmitted, _ := global.Rdb.SIsMember(global.Rctx, "teams", teamID).Result()

	// 返回结果
	utility.ResponseSuccess(context, gin.H{
		"id":          person.TeamId,
		"name":        team.Name,
		"route":       team.Route,
		"password":    team.Password,
		"submitted":   teamSubmitted,
		"allow_match": team.AllowMatch,
		"slogan":      team.Slogan,
		"point":       team.Point,
		"status":      team.Status,
		"leader":      captainData,
		"member":      memberData,
	})
}
