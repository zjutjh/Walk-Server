package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
)

type UpdateTeamRequest struct {
	Name       string `json:"name" binding:"required"`
	Slogan     string `json:"slogan"`
	Route      uint8  `json:"route" binding:"required"`
	AllowMatch *bool  `json:"allow_match" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func UpdateTeamHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateTeamRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

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

		if person.Status != 2 {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以修改队伍信息"))
			return
		}

		var team model.Team
		if err := db.First(&team, person.TeamId).Error; err != nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		// Check name uniqueness if changed
		if team.Name != req.Name {
			var count int64
			db.Model(&model.Team{}).Where("name = ? AND id != ?", req.Name, team.ID).Count(&count)
			if count > 0 {
				reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍名已存在"))
				return
			}
		}

		team.Name = req.Name
		team.Slogan = req.Slogan
		team.Route = req.Route
		team.AllowMatch = *req.AllowMatch
		team.Password = req.Password

		if err := db.Save(&team).Error; err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
