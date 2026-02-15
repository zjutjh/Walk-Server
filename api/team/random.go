package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"gorm.io/gorm"
)

func GetRandomListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		teamRepo := repo.NewTeamRepo()
		teams, err := teamRepo.GetRandomList(c.Request.Context(), 10)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, teams)
	}
}

type JoinRandomRequest struct {
	TeamID int64 `json:"team_id" binding:"required"`
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
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "已加入队伍"))
			return
		}

		team, err := teamRepo.FindById(c.Request.Context(), req.TeamID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if team == nil {
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

		err = teamRepo.Transaction(c.Request.Context(), func(tx *gorm.DB) error {
			t, err := teamRepo.FindByIdForUpdate(c.Request.Context(), tx, req.TeamID)
			if err != nil {
				return err
			}
			if t.Num >= 6 {
				return gorm.ErrInvalidData
			}

			person.TeamID = &t.ID
			person.Status = comm.PersonStatusMember
			if err := personRepo.Update(c.Request.Context(), tx, person); err != nil {
				return err
			}

			t.Num++
			if err := teamRepo.Save(c.Request.Context(), tx, t); err != nil {
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
