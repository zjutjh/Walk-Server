package team

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
)

type UpdateTeamRequest struct {
	Name       string `json:"name" binding:"required"`
	Slogan     string `json:"slogan"`
	Route      uint8  `json:"route" binding:"required"`
	AllowMatch *bool  `json:"allow_match" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

func UpdateTeamHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateTeamRequest
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

		if person.TeamId == 0 {
			reply.Fail(c, comm.WithMsg(comm.CodeDataNotFound, "未加入队伍"))
			return
		}

		if person.Status != comm.PersonStatusCaptain {
			reply.Fail(c, comm.WithMsg(comm.CodePermissionDenied, "只有队长可以修改队伍信息"))
			return
		}

		team, err := teamRepo.FindById(c.Request.Context(), person.TeamId)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if team == nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		// 如果修改了名称，检查名称是否重复
		if team.Name != req.Name {
			exists, err := teamRepo.CheckNameExistsExcludingId(c.Request.Context(), req.Name, team.ID)
			if err != nil {
				reply.Fail(c, comm.CodeDatabaseError)
				return
			}
			if exists {
				reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "队伍名已存在"))
				return
			}
		}

		team.Name = req.Name
		team.Slogan = req.Slogan
		team.Route = req.Route
		team.AllowMatch = *req.AllowMatch
		team.Password = req.Password

		if err := teamRepo.Save(c.Request.Context(), nil, team); err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
