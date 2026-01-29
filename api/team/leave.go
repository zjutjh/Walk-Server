package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"gorm.io/gorm"
)

func LeaveTeamHandler() gin.HandlerFunc {
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

		if person.Status == comm.PersonStatusCaptain {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "队长不能退出队伍，请先解散或转让队长"))
			return
		}

		err = teamRepo.Transaction(c.Request.Context(), func(tx *gorm.DB) error {
			// 更新人员信息
			oldTeamID := person.TeamId
			person.TeamId = 0
			person.Status = comm.PersonStatusNone
			if err := personRepo.Update(c.Request.Context(), tx, person); err != nil {
				return err
			}

			// 更新队伍人数
			team, err := teamRepo.FindById(c.Request.Context(), oldTeamID)
			if err != nil {
				return err
			}
			team.Num--
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
