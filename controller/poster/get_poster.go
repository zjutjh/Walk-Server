package poster

import (
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

var teamRouteMap = map[uint8]string{
	1: "朝晖全程",
	2: "屏峰半程",
	3: "屏峰全程",
	4: "莫干山半程",
	5: "莫干山全程",
}

func GetPoster(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 获取团队信息
	user, _ := model.GetPerson(jwtData.OpenID)
	team, err := model.GetTeamInfo(uint(user.TeamId))
	if err != nil {
		utility.ResponseError(context, "no team")
		return
	}

	// 获取团队成员
	captain, members := model.GetPersonsInTeam(user.TeamId)
	var memberNames []string
	memberNames = append(memberNames, captain.Name)
	for _, member := range members {
		memberNames = append(memberNames, member.Name)
	}

	imgUrl, err := utility.Poster(teamRouteMap[team.Route], team.Name, team.Slogan, int(team.Num), memberNames)
	if err != nil {
		utility.ResponseError(context, "海报生成错误")
		return
	}

	utility.ResponseSuccess(context, gin.H{
		"img_url": imgUrl,
	})
}
