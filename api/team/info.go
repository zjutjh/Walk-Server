package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
)

func GetTeamInfoHandler() gin.HandlerFunc {
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

		// Construct response similar to main branch
		var memberData []gin.H
		for _, m := range members {
			memberData = append(memberData, gin.H{
				"name":    m.Name,
				"gender":  m.Gender,
				"open_id": m.OpenId,
				"campus":  m.Campus,
				"type":    m.Type,
				"contact": gin.H{
					"qq":     m.Qq,
					"wechat": m.Wechat,
					"tel":    m.Tel,
				},
				"walk_status": m.WalkStatus,
			})
		}

		reply.Success(c, gin.H{
			"id":          team.ID,
			"name":        team.Name,
			"route":       team.Route,
			"password":    team.Password,
			"allow_match": team.AllowMatch,
			"slogan":      team.Slogan,
			"point":       team.Point,
			"status":      team.Status,
			"start_num":   team.StartNum,
			"code":        team.Code,
			"submit":      team.Submit,
			"members":     memberData,
		})
	}
}
