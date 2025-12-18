package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

func GetRandomListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := ndb.Pick()
		var teams []model.Team
		// Find teams that allow match and are not full (assuming max 6)
		if err := db.Where("allow_match = ? AND num < ?", true, 6).Limit(10).Find(&teams).Error; err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, teams)
	}
}

type JoinRandomRequest struct {
	TeamID uint `json:"team_id" binding:"required"`
}

func JoinRandomHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req JoinRandomRequest
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

		if person.TeamId != 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "已加入队伍"))
			return
		}

		var team model.Team
		if err := db.First(&team, req.TeamID).Error; err != nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "队伍不存在"))
			return
		}

		if !team.AllowMatch {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "该队伍不允许随机匹配"))
			return
		}

		if team.Num >= 6 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍已满"))
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			// Double check lock? For now simple transaction
			// Re-read team in transaction to ensure consistency?
			var t model.Team
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&t, req.TeamID).Error; err != nil {
				return err
			}
			if t.Num >= 6 {
				return gorm.ErrInvalidData // Use a standard error or custom one
			}

			person.TeamId = t.ID
			person.Status = 1 // Member
			if err := tx.Save(&person).Error; err != nil {
				return err
			}

			t.Num++
			if err := tx.Save(&t).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			if err == gorm.ErrInvalidData {
				reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍已满"))
				return
			}
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
