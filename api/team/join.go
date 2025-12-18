package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type JoinTeamRequest struct {
	TeamID   int64  `json:"team_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func JoinTeamHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		var req JoinTeamRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		db := ndb.Pick()
		var person model.Person
		if err := db.Where("open_id = ?", openID).First(&person).Error; err != nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		if person.TeamId != 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "已加入队伍"))
			return
		}

		var team model.Team
		if err := db.First(&team, req.TeamID).Error; err != nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "队伍不存在"))
			return
		}

		if team.Password != req.Password {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "密码错误"))
			return
		}

		// Check team size limit (assuming 6 based on typical rules, or check config)
		if team.Num >= 6 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍已满"))
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			person.TeamId = team.ID
			person.Status = 1 // Member
			if err := tx.Save(&person).Error; err != nil {
				return err
			}
			team.Num++
			if err := tx.Save(&team).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
