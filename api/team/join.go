package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"gorm.io/gorm"
)

type JoinTeamRequest struct {
	TeamID   int64  `json:"team_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func JoinTeamHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		var req JoinTeamRequest
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

		if person.TeamId != 0 {
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

		if team.Password != req.Password {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "密码错误"))
			return
		}

		// 检查队伍人数限制（假设规则为6人，或者查看配置）
		if team.Num >= comm.MaxTeamMember {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍已满"))
			return
		}

		err = teamRepo.Transaction(c.Request.Context(), func(tx *gorm.DB) error {
			person.TeamId = team.ID
			person.Status = comm.PersonStatusMember
			if err := personRepo.Update(c.Request.Context(), tx, person); err != nil {
				return err
			}
			team.Num++
			if err := teamRepo.Update(c.Request.Context(), tx, team); err != nil {
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
