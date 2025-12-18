package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type RemoveMemberRequest struct {
	MemberID uint `json:"member_id" binding:"required"`
}

func RemoveMemberHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RemoveMemberRequest
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
		var captain model.Person
		if err := db.Where("open_id = ?", openID).First(&captain).Error; err != nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		if captain.TeamId == 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "未加入队伍"))
			return
		}

		if captain.Status != 2 {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以移除队员"))
			return
		}

		var target model.Person
		if err := db.First(&target, req.MemberID).Error; err != nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "队员不存在"))
			return
		}

		if target.TeamId != captain.TeamId {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "该队员不在你的队伍中"))
			return
		}

		if target.ID == captain.ID {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "不能移除自己"))
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			// Update target
			target.TeamId = 0
			target.Status = 0
			if err := tx.Save(&target).Error; err != nil {
				return err
			}

			// Update team count
			var team model.Team
			if err := tx.First(&team, captain.TeamId).Error; err != nil {
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
