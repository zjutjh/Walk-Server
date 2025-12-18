package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

func DisbandTeamHandler() gin.HandlerFunc {
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

		if person.Status != 2 {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以解散队伍"))
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			// Reset all members
			if err := tx.Model(&model.Person{}).Where("team_id = ?", person.TeamId).Updates(map[string]interface{}{"team_id": 0, "status": 0}).Error; err != nil {
				return err
			}

			// Delete team
			if err := tx.Delete(&model.Team{}, person.TeamId).Error; err != nil {
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
