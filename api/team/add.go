package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
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
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以添加队员"))
			return
		}

		team, err := teamRepo.FindById(c.Request.Context(), *captain.TeamID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if team == nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		if team.Num >= 6 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍已满"))
			return
		}

		target, err := personRepo.FindByStuId(c.Request.Context(), req.StuID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if target == nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "该学号未报名"))
			return
		}

		if target.TeamID != nil && *target.TeamID > 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "该同学已加入其他队伍"))
			return
		}

		err = teamRepo.Transaction(c.Request.Context(), func(tx *gorm.DB) error {
			// 更新目标队员
			target.TeamID = &team.ID
			target.Status = comm.PersonStatusMember
			if err := personRepo.Update(c.Request.Context(), tx, target); err != nil {
				return err
			}

			// 更新队伍人数
			team.Num++
			if err := teamRepo.Save(c.Request.Context(), tx, team); err != nil {
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
