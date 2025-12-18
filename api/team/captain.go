package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type TransferCaptainRequest struct {
	MemberID uint `json:"member_id" binding:"required"`
}

func TransferCaptainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TransferCaptainRequest
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
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以转让队长"))
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
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "不能转让给自己"))
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			// Old captain -> Member
			captain.Status = 1
			if err := tx.Save(&captain).Error; err != nil {
				return err
			}

			// New captain -> Captain
			target.Status = 2
			if err := tx.Save(&target).Error; err != nil {
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
