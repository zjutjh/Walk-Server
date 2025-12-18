package team

import (
	"app/comm"
	"app/dao/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type CreateTeamRequest struct {
	Name       string `json:"name" binding:"required"`
	Route      uint8  `json:"route" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Slogan     string `json:"slogan" binding:"required"`
	AllowMatch *bool  `json:"allow_match" binding:"required"`
}

func CreateTeamHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		var req CreateTeamRequest
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
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "已有队伍"))
			return
		}

		// Check team name unique
		var count int64
		db.Model(&model.Team{}).Where("name = ?", req.Name).Count(&count)
		if count > 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍名已存在"))
			return
		}

		team := model.Team{
			Name:       req.Name,
			Route:      req.Route,
			Password:   req.Password,
			Slogan:     req.Slogan,
			AllowMatch: *req.AllowMatch,
			Num:        1,
			Status:     1, // Assuming 1 is active/created
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&team).Error; err != nil {
				return err
			}
			person.TeamId = team.ID
			person.Status = 2 // Captain
			if err := tx.Save(&person).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, gin.H{"team_id": team.ID})
	}
}
