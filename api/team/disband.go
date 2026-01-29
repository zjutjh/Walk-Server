package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"gorm.io/gorm"
)

func DisbandTeamHandler() gin.HandlerFunc {
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

		if person.TeamId == 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "未加入队伍"))
			return
		}

		if person.Status != comm.PersonStatusCaptain {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以解散队伍"))
			return
		}

		err = teamRepo.Transaction(c.Request.Context(), func(tx *gorm.DB) error {
			// Reset all members
			if err := personRepo.ResetTeamInfo(c.Request.Context(), tx, person.TeamId); err != nil {
				return err
			}

			// Delete team
			if err := teamRepo.Delete(c.Request.Context(), tx, person.TeamId); err != nil {
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
