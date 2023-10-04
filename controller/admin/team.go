package admin

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"walk-server/constant"
	"walk-server/global"
	"walk-server/middleware"
	"walk-server/model"
	"walk-server/service/adminService"
	"walk-server/service/teamService"
	"walk-server/service/userService"
	"walk-server/utility"
)

type TeamForm struct {
	TeamID uint `json:"team_id" binding:"required"`
}

func GetTeam(c *gin.Context) {
	TeamID, err := strconv.Atoi(c.Query("team_id"))

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	user, err := adminService.GetAdminByJWT(c)
	team, err := teamService.GetTeamByID(uint(TeamID))
	if team == nil || err != nil {
		utility.ResponseError(c, "服务错误")
		return
	}

	b := middleware.CheckRoute(user, team)
	if !b {
		utility.ResponseError(c, "管理员权限不足")
		return
	}

	var persons []model.Person
	global.DB.Where("team_id = ?", team.ID).Find(&persons)

	var memberData []gin.H
	for _, member := range persons {
		memberData = append(memberData, gin.H{
			"name":    member.Name,
			"gender":  member.Gender,
			"open_id": member.OpenId,
			"campus":  member.Campus,
			"contact": gin.H{
				"qq":     member.Qq,
				"wechat": member.Wechat,
				"tel":    member.Tel,
			},
			"walk_status": member.WalkStatus,
		})
	}
	utility.ResponseSuccess(c, gin.H{
		"id":          TeamID,
		"name":        team.Name,
		"route":       team.Route,
		"password":    team.Password,
		"allow_match": team.AllowMatch,
		"slogan":      team.Slogan,
		"point":       team.Point,
		"status":      team.Status,
		"start_num":   team.StartNum,
		"member":      memberData,
	})
}

// TeamSM 团队扫码
func TeamSM(c *gin.Context) {
	var postForm TeamForm
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	user, err := adminService.GetAdminByJWT(c)
	team, err := teamService.GetTeamByID(postForm.TeamID)
	if team == nil || err != nil {
		utility.ResponseError(c, "服务错误")
		return
	}

	b := middleware.CheckRoute(user, team)
	if !b {
		utility.ResponseError(c, "管理员权限不足")
		return
	}

	if team.Status == 3 || team.Status == 4 {
		utility.ResponseError(c, "团队已结束毅行")
		return
	}

	team.Status = 5
	teamService.Update(*team)
	utility.ResponseSuccess(c, nil)
}

func UpdateTeam(c *gin.Context) {
	var postForm TeamForm
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	user, err := adminService.GetAdminByJWT(c)
	team, err := teamService.GetTeamByID(postForm.TeamID)
	if team == nil || err != nil {
		utility.ResponseError(c, "服务错误")
		return
	}

	b := middleware.CheckRoute(user, team)
	if !b {
		utility.ResponseError(c, "管理员权限不足")
		return
	}

	if team.Status != 5 {
		utility.ResponseError(c, "团队未扫码")
		return
	}
	var persons []model.Person
	global.DB.Where("team_id = ?", team.ID).Find(&persons)
	flag := true
	var num uint
	num = 0
	for _, p := range persons {
		if p.WalkStatus != 3 && p.WalkStatus != 4 {
			flag = false
			break
		} else {
			if p.WalkStatus == 3 {
				num++
			}
		}
	}

	if !flag {
		utility.ResponseError(c, "还有成员未扫码")
		return
	}

	if num == 0 {
		team.Status = 3
		teamService.Update(*team)
		utility.ResponseSuccess(c, gin.H{
			"progress_num": 0,
		})
		return
	}

	team.Point++

	switch team.Point {
	case constant.PointMap[team.Route]:
		{
			for _, p := range persons {
				if p.WalkStatus == 3 {
					p.WalkStatus = 5
					userService.Update(p)
				}
			}
			flagNum := team.StartNum / 2
			if flagNum > num {
				team.Status = 3
			} else {
				team.Status = 4
			}
			teamService.Update(*team)
			utility.ResponseSuccess(c, gin.H{
				"progress_num": 0,
			})
			return
		}
	case 1:
		{
			team.StartNum = num
			teamService.Update(*team)
		}
	}
	for _, p := range persons {
		if p.WalkStatus == 3 {
			p.WalkStatus = 2
			userService.Update(p)
		}
	}
	team.Status = 2
	teamService.Update(*team)
	utility.ResponseSuccess(c, gin.H{
		"progress_num": num,
	})
	return
}
