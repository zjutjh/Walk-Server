package poster

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
)

var teamRouteMap = map[uint8]string{
	1: "朝晖全程",
	2: "屏峰半程",
	3: "屏峰全程",
	4: "莫干山半程",
	5: "莫干山全程",
}

func GetPosterHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		db := ndb.Pick()
		var person model.Person
		if err := db.Where("open_id = ?", openID).First(&person).Error; err != nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		if person.TeamId == 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "未加入队伍"))
			return
		}

		var team model.Team
		if err := db.First(&team, person.TeamId).Error; err != nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		var members []model.Person
		if err := db.Where("team_id = ?", team.ID).Find(&members).Error; err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		var memberNames []string
		for _, m := range members {
			memberNames = append(memberNames, m.Name)
		}

		imgUrl, err := comm.GeneratePoster(teamRouteMap[team.Route], team.Name, team.Slogan, team.Num, memberNames)
		if err != nil {
			reply.Fail(c, comm.WithMsg(comm.CodeThirdServiceError, err.Error()))
			return
		}

		reply.Success(c, gin.H{"img_url": imgUrl})
	}
}
