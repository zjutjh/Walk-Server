package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
)

func GetTeamInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		personRepo := repo.NewPersonRepo()
		teamRepo := repo.NewTeamRepo()

		person, err := personRepo.FindByOpenId(c.Request.Context(), openID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if person == nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		if person.TeamID == nil || *person.TeamID <= 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "未加入队伍"))
			return
		}

		team, err := teamRepo.FindById(c.Request.Context(), *person.TeamID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if team == nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		members, err := personRepo.FindByTeamId(c.Request.Context(), team.ID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		// 构建与主分支类似的响应
		var memberData []gin.H
		for _, m := range members {
			memberData = append(memberData, gin.H{
				"name":    m.Name,
				"gender":  m.Gender,
				"open_id": m.OpenID,
				"campus":  m.Campus,
				"type":    m.Type,
				"contact": gin.H{
					"qq":     m.QQ,
					"wechat": m.Wechat,
					"tel":    m.Tel,
				},
				"walk_status": m.WalkStatus,
			})
		}

		reply.Success(c, gin.H{
			"id":           team.ID,
			"name":         team.Name,
			"route":        team.Route,
			"password":     team.Password,
			"allow_match":  team.AllowMatch,
			"slogan":       team.Slogan,
			"point_id":     team.PointID,
			"num":          team.Num,
			"is_departed":  team.IsDeparted,
			"is_completed": team.IsCompleted,
			"code":         team.Code,
			"submit":       team.Submit,
			"members":      memberData,
		})
	}
}
