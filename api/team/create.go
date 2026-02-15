package team

import (
	"app/comm"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"gorm.io/gorm"
)

type CreateTeamRequest struct {
	Name       string `json:"name" binding:"required"`
	Route      string `json:"route" binding:"required"`
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

		if person.TeamID != nil && *person.TeamID > 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "已有队伍"))
			return
		}

		// 检查队伍名唯一性
		exists, err := teamRepo.CheckNameExistsExcludingId(c.Request.Context(), req.Name, 0)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if exists {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍名已存在"))
			return
		}

		team := model.Team{
			Name:       req.Name,
			Route:      req.Route,
			Password:   req.Password,
			Slogan:     &req.Slogan,
			AllowMatch: *req.AllowMatch,
			Captain:    openID,
			Num:        1,
		}

		err = teamRepo.Transaction(c.Request.Context(), func(tx *gorm.DB) error {
			if err := teamRepo.Create(c.Request.Context(), tx, &team); err != nil {
				return err
			}
			person.TeamID = &team.ID
			person.Status = comm.PersonStatusCaptain
			if err := personRepo.Update(c.Request.Context(), tx, person); err != nil {
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
