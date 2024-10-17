package admin

import (
	"strconv"
	"walk-server/constant"
	"walk-server/global"
	"walk-server/middleware"
	"walk-server/model"
	"walk-server/service/adminService"
	"walk-server/service/teamService"
	"walk-server/service/userService"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeamForm struct {
	CodeType uint   `form:"code_type" binding:"required,oneof=1 2"` // 1团队码2签到码
	Content  string `form:"content" binding:"required"`             // 团队码为team_id，签到码为code
}

func GetTeam(c *gin.Context) {
	var postForm TeamForm
	err := c.ShouldBindQuery(&postForm)
	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	user, _ := adminService.GetAdminByJWT(c)
	var team *model.Team
	if postForm.CodeType == 1 {
		teamID, convErr := strconv.ParseUint(postForm.Content, 10, 32)
		if convErr != nil {
			utility.ResponseError(c, "参数错误")
			return
		}
		team, err = teamService.GetTeamByID(uint(teamID))
	} else {
		team, err = teamService.GetTeamByCode(postForm.Content)
	}
	if team == nil || err != nil {
		utility.ResponseError(c, "二维码错误，队伍查找失败")
		return
	}

	b := middleware.CheckRoute(user, team)
	if !b {
		utility.ResponseError(c, "该队伍为其他路线")
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
			"type":    member.Type,
			"contact": gin.H{
				"qq":     member.Qq,
				"wechat": member.Wechat,
				"tel":    member.Tel,
			},
			"walk_status": member.WalkStatus,
		})
	}
	utility.ResponseSuccess(c, gin.H{
		"team": gin.H{
			"id":          team.ID,
			"name":        team.Name,
			"route":       team.Route,
			"password":    team.Password,
			"allow_match": team.AllowMatch,
			"slogan":      team.Slogan,
			"point":       team.Point,
			"status":      team.Status,
			"start_num":   team.StartNum,
			"code":        team.Code,
		},
		"admin":  user,
		"member": memberData,
	})
}

type BindTeamForm struct {
	TeamID uint   `json:"team_id" binding:"required"`
	Type   uint   `json:"type" binding:"required,eq=2"`
	Code   string `json:"code" binding:"required"`
}

func BindTeam(c *gin.Context) {
	var postForm BindTeamForm
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	user, _ := adminService.GetAdminByJWT(c)
	team, err := teamService.GetTeamByID(postForm.TeamID)
	if team == nil || err != nil {
		utility.ResponseError(c, "队伍查找失败，请重新核对")
		return
	}

	b := middleware.CheckRoute(user, team)
	if !b {
		utility.ResponseError(c, "该队伍为其他路线")
		return
	}

	_, err = teamService.GetTeamByCode(postForm.Code)
	if err == nil {
		utility.ResponseError(c, "二维码已绑定")
		return
	} else if err != gorm.ErrRecordNotFound {
		utility.ResponseError(c, "服务错误")
		return
	}
	var persons []model.Person
	global.DB.Where("team_id = ?", team.ID).Find(&persons)
	flag := true
	num := uint(0)
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
		utility.ResponseError(c, "还有成员未确认状态")
		return
	}

	if (team.Num+1)/2 > uint8(num) {
		utility.ResponseError(c, "团队人数不足，无法绑定")
		return
	}

	team.Code = postForm.Code
	team.Status = 5
	team.StartNum = num
	teamService.Update(*team)
	utility.ResponseSuccess(c, nil)
}

type TeamStatusForm struct {
	CodeType uint   `json:"code_type" binding:"required"` //1团队码2签到码
	Content  string `json:"content" binding:"required"`   //团队码为team_id，签到码为code
}

func UpdateTeamStatus(c *gin.Context) {
	var postForm TeamStatusForm
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	user, _ := adminService.GetAdminByJWT(c)
	var team *model.Team
	if postForm.CodeType == 1 {
		teamID, convErr := strconv.ParseUint(postForm.Content, 10, 32)
		if convErr != nil {
			utility.ResponseError(c, "参数错误")
			return
		}
		team, err = teamService.GetTeamByID(uint(teamID))
	} else if postForm.CodeType == 2 {
		team, err = teamService.GetTeamByCode(postForm.Content)
	} else {
		utility.ResponseError(c, "参数错误")
		return
	}

	if team == nil || err != nil {
		utility.ResponseError(c, "队伍查找失败，请重新核对")
		return
	}

	b := middleware.CheckRoute(user, team)
	if !b {
		utility.ResponseError(c, "该队伍为其他路线")
		return
	}
	if team.Status != 5 && team.Status != 2 {
		utility.ResponseError(c, "团队起点未扫码")
		return
	}
	var persons []model.Person
	global.DB.Where("team_id = ?", team.ID).Find(&persons)
	num := uint(0)
	for _, p := range persons {
		if p.WalkStatus == 3 || p.WalkStatus == 2 {
			num++
		}
	}

	if num == 0 {
		team.Status = 3
		team.Point = int8(constant.PointMap[team.Route])
		teamService.Update(*team)
		utility.ResponseSuccess(c, gin.H{
			"progress_num": 0,
		})
		return
	}

	// 各路线点位签到逻辑设置
	switch team.Route {
	case 5:
		switch {
		case team.Point == 1 && (user.Point == 2 || user.Point == 6):
			team.Point = 2
		case team.Point == 4 && (user.Point == 2 || user.Point == 6):
			team.Point = 6
		default:
			team.Point = user.Point
		}
	case 2:
		if user.Point > 2 {
			team.Point = user.Point - 2
		}
	case 3:
		if user.Route == 2 {
			utility.ResponseError(c, "该队伍为其他路线")
			return
		}
	default:
		team.Point = user.Point
	}

	if team.Point == int8(constant.PointMap[team.Route]) {
		for _, p := range persons {
			if p.WalkStatus == 2 {
				p.WalkStatus = 5
				userService.Update(p)
			}
		}
		team.Status = 4
		teamService.Update(*team)
		utility.ResponseSuccess(c, gin.H{
			"progress_num": 0,
		})
		return
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
}

type RegroupForm struct {
	Jwts   []string `json:"jwts" binding:"required"`
	Secret string   `json:"secret" binding:"required"`
	Route  uint8    `json:"route" binding:"required"`
}

func Regroup(c *gin.Context) {
	var postForm RegroupForm
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}

	var persons []model.Person
	processedJwts := make(map[string]bool)
	for _, jwt := range postForm.Jwts {
		if processedJwts[jwt] {
			utility.ResponseError(c, "重复扫码,请重新提交")
		}
		processedJwts[jwt] = true

		jwtToken := jwt[7:]
		jwtData, err := utility.ParseToken(jwtToken)

		if err != nil {
			utility.ResponseError(c, "扫码错误，请重新扫码")
			return
		}

		// 获取个人信息
		person, err := model.GetPerson(jwtData.OpenID)

		if err != nil {
			utility.ResponseError(c, "扫码错误，请重新扫码")
			return
		}

		// 如果已有队伍则退出
		if person.TeamId != -1 {
			captain, persons := model.GetPersonsInTeam(person.TeamId)
			for _, p := range persons {
				p.TeamId = -1
				p.Status = 0
				p.WalkStatus = 1
				userService.Update(p)
			}
			captain.TeamId = -1
			captain.Status = 0
			captain.WalkStatus = 1
			userService.Update(captain)
			team, _ := teamService.GetTeamByID(uint(person.TeamId))
			err = teamService.Delete(*team)
			if err != nil {
				utility.ResponseError(c, "服务错误")
				return
			}
		}

		persons = append(persons, *person)
	}

	// 创建新队伍，第一个人作为队长
	newTeam := model.Team{
		Name:       "新队伍",
		Route:      postForm.Route,
		Password:   "123456",
		AllowMatch: true,
		Slogan:     "新的开始",
		Point:      0,
		Status:     1,
		StartNum:   uint(len(persons)),
		Num:        uint8(len(persons)),
		Captain:    persons[0].OpenId,
		Submit:     true,
	}
	teamService.Create(newTeam)

	team, err := teamService.GetTeamByCaptain(persons[0].OpenId)
	if err != nil {
		utility.ResponseError(c, "服务错误")
		return
	}

	// 更新每个人的队伍ID
	for i, person := range persons {
		person.TeamId = int(team.ID)
		if i == 0 {
			person.Status = 2
		} else {
			person.Status = 1
		}
		userService.Update(person)
	}
	global.Rdb.SAdd(global.Rctx, "teams", strconv.Itoa(int(newTeam.ID)))

	utility.ResponseSuccess(c, gin.H{
		"team_id": newTeam.ID,
	})
}

type SubmitTeamForm struct {
	TeamID uint   `json:"team_id" binding:"required"`
	Secret string `json:"secret" binding:"required"`
}

func SubmitTeam(c *gin.Context) {
	var postForm SubmitTeamForm
	err := c.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}
	team, err := teamService.GetTeamByID(postForm.TeamID)
	if team == nil || err != nil {
		utility.ResponseError(c, "队伍查找失败，请重新核对")
		return
	}

	team.Submit = true
	teamService.Update(*team)
	global.Rdb.SAdd(global.Rctx, "teams", strconv.Itoa(int(team.ID)))
	utility.ResponseSuccess(c, nil)

}

type GetDetailForm struct {
	Secret string `form:"secret" binding:"required"`
}

// GetDetail 获取全部路线的点位信息
func GetDetail(c *gin.Context) {
	var postForm GetDetailForm
	if err := c.ShouldBindQuery(&postForm); err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}
	if postForm.Secret != global.Config.GetString("server.secret") {
		utility.ResponseError(c, "密码错误")
		return
	}

	routes := map[string]int{
		"zh":      1,
		"pfHalf":  2,
		"pfAll":   3,
		"mgsHalf": 4,
		"mgsAll":  5,
	}

	resultMap := make(map[string][]int64)
	for key, route := range routes {
		resultMap[key] = make([]int64, constant.PointMap[uint8(route)]+3)
	}

	// 获取各点位人数
	getPointCounts := func(route int, status []int, team_stuats []int, points []int64) {
		var pointCounts []struct {
			Point int64
			Count int64
		}
		global.DB.Model(&model.Person{}).
			Select("teams.point, count(*) as count").
			Joins("JOIN teams ON people.team_id = teams.id").
			Where("teams.route = ? AND people.walk_status IN ? AND teams.status IN ?", route, status, team_stuats).
			Group("teams.point").
			Order("teams.point").
			Scan(&pointCounts)

		for _, pointCount := range pointCounts {
			if pointCount.Point >= 0 && int(pointCount.Point) < int(constant.PointMap[uint8(route)])+1 {
				points[pointCount.Point+1] = pointCount.Count
			}
		}

	}

	// 获取各路线未开始人数
	getStartCounts := func(route int, points *int64) {
		global.DB.Model(&model.Person{}).
			Select("count(*) as count").
			Joins("JOIN teams ON people.team_id = teams.id").
			Where("teams.route = ? AND people.walk_status = 1 And teams.submit = 1", route).
			Pluck("count", points)
	}

	// 获取各路线已结束和下撤人数
	appendEndCounts := func(route int, points []int64) {
		var endCount5, endCount4 int64
		global.DB.Model(&model.Person{}).
			Select("count(*) as count").
			Joins("JOIN teams ON people.team_id = teams.id").
			Where("teams.route = ? AND people.walk_status = 5", route).
			Pluck("count", &endCount5)
		global.DB.Model(&model.Person{}).
			Select("count(*) as count").
			Joins("JOIN teams ON people.team_id = teams.id").
			Where("teams.route = ? AND people.walk_status = 4", route).
			Pluck("count", &endCount4)
		points[len(points)-2] = endCount5
		points[len(points)-1] = endCount4
	}

	// 状态：进行中、未开始、已结束
	personStatusInProgress := []int{2, 3}
	teamStatusInProgress := []int{2, 5}

	for key, route := range routes {
		getPointCounts(route, personStatusInProgress, teamStatusInProgress, resultMap[key])
		getStartCounts(route, &resultMap[key][0])
		appendEndCounts(route, resultMap[key])
	}

	// 返回结果
	utility.ResponseSuccess(c, gin.H{
		"zh":      resultMap["zh"],
		"pfAll":   resultMap["pfAll"],
		"pfHalf":  resultMap["pfHalf"],
		"mgsHalf": resultMap["mgsHalf"],
		"mgsAll":  resultMap["mgsAll"],
	})
}
