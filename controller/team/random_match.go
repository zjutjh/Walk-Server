package team

import (
	"math/rand"
	"strconv"
	"time"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// RandomMatch 随机组队
func RandomMatch(context *gin.Context) {
	route, err := strconv.Atoi(context.Query("route"))
	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}

	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 从数据库中读取用户信息
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status != 0 { // 如果在一个团队中
		utility.ResponseError(context, "请退出或解散原来的团队")
		return
	}

	if person.JoinOp == 0 { // 加入次数用完了
		utility.ResponseError(context, "没有加入次数了")
		return
	}

	// 找到一个可以加入的随机队伍
	var teams, randTeams []model.Team
	global.DB.Where("num < 6 AND route = ? AND allow_match = 1 AND submitted = 0", route).Find(&teams) // 从可能的队伍中挑选前 10 条
	rand.Seed(time.Now().UnixNano())                                                                    // 设置随机种子
	for n := len(teams); n > 0; n-- {
		randIndex := rand.Intn(n)
		randTeams = append(randTeams, teams[randIndex])
		teams[n-1], teams[randIndex] = teams[randIndex], teams[n-1] // 将 teams[randIndex] 丢到末尾
	}
	for _, team := range randTeams {
		result := global.DB.Model(&team).Where("num < 6 AND allow_match = 1 AND submitted = 0").Update("num", team.Num+1) // 更新队伍人数
		if result.RowsAffected != 0 {                                                                                      // 更新成功
			person.Status = 1
			person.JoinOp--
			person.TeamId = int(team.ID)
			model.UpdatePerson(jwtData.OpenID, person) // 将新的用户信息写入数据库
			utility.ResponseSuccess(context, nil)

			return
		}
	}

	utility.ResponseError(context, "没有匹配上的队伍")
}