package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"gorm.io/gorm"
)

type RemoveMemberRequest struct {
	MemberID int64 `json:"member_id" binding:"required"`
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

		personRepo := repo.NewPersonRepo()
		teamRepo := repo.NewTeamRepo()

		captain, err := personRepo.FindByOpenId(c.Request.Context(), openID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if captain == nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		if captain.TeamId == 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "未加入队伍"))
			return
		}

		if captain.Status != comm.PersonStatusCaptain {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以移除队员"))
			return
		}

		target, err := personRepo.FindById(c.Request.Context(), req.MemberID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if target == nil {
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

		err = teamRepo.Transaction(c.Request.Context(), func(tx *gorm.DB) error {
			// 更新目标队员
			target.TeamId = 0
			target.Status = comm.PersonStatusNone
			if err := personRepo.Update(c.Request.Context(), tx, target); err != nil {
				return err
			}

			// 更新队伍人数
			team, err := teamRepo.FindById(c.Request.Context(), captain.TeamId)
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
