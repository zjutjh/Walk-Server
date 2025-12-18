package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

func LeaveTeamHandler() gin.HandlerFunc {
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

		if person.Status == 2 {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "队长不能退出队伍，请先解散或转让队长"))
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			// Update person
			oldTeamID := person.TeamId
			person.TeamId = 0
			person.Status = 0
			if err := tx.Save(&person).Error; err != nil {
				return err
			}

			// Update team count
			var team model.Team
			if err := tx.First(&team, oldTeamID).Error; err != nil {
				return err
			}
			team.Num--
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
