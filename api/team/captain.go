package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"gorm.io/gorm"
)

type TransferCaptainRequest struct {
	MemberOpenID string `json:"member_open_id" binding:"required"`
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

		if captain.TeamID == nil || *captain.TeamID <= 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "未加入队伍"))
			return
		}

		if captain.Status != comm.PersonStatusCaptain {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以转让队长"))
			return
		}

		target, err := personRepo.FindByOpenId(c.Request.Context(), req.MemberOpenID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if target == nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "队员不存在"))
			return
		}

		if target.TeamID == nil || *target.TeamID != *captain.TeamID {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "该队员不在你的队伍中"))
			return
		}

		if target.OpenID == captain.OpenID {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "不能转让给自己"))
			return
		}

		err = teamRepo.Transaction(c.Request.Context(), func(tx *gorm.DB) error {
			// Old captain -> Member
			captain.Status = comm.PersonStatusMember
			if err := personRepo.Update(c.Request.Context(), tx, captain); err != nil {
				return err
			}

			// New captain -> Captain
			target.Status = comm.PersonStatusCaptain
			if err := personRepo.Update(c.Request.Context(), tx, target); err != nil {
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
