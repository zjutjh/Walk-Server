package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type AddMemberRequest struct {
	StuID string `json:"stu_id" binding:"required"`
}

func AddMemberHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddMemberRequest
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
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以添加队员"))
			return
		}

		var team model.Team
		if err := db.First(&team, captain.TeamId).Error; err != nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		if team.Num >= 6 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍已满"))
			return
		}

		var target model.Person
		if err := db.Where("stu_id = ?", req.StuID).First(&target).Error; err != nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "该学号未报名"))
			return
		}

		if target.TeamId != 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "该同学已加入其他队伍"))
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			// Update target
			target.TeamId = team.ID
			target.Status = 1 // Member
			if err := tx.Save(&target).Error; err != nil {
				return err
			}

			// Update team count
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
