package admin

import (
	"github.com/gin-gonic/gin"
	"walk-server/global"
	"walk-server/middleware"
	"walk-server/model"
	"walk-server/service/adminService"
	"walk-server/service/userService"
	"walk-server/utility"
)

type UserSMForm struct {
	Jwt        string `json:"user_jwt" binding:"required"`
	WalkStatus uint   `json:"walk_status" binding:"required"`
}

func UserSM(c *gin.Context) {
	var postForm UserSMForm
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	user, err := adminService.GetAdminByJWT(c)

	jwtToken := postForm.Jwt
	jwtToken = jwtToken[7:]
	jwtData, err := utility.ParseToken(jwtToken)

	if err != nil {
		utility.ResponseError(c, "扫码错误")
	}

	// 获取个人信息
	person, err := model.GetPerson(jwtData.OpenID)

	if err != nil {
		utility.ResponseError(c, "扫码错误")
	}

	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)

	b := middleware.CheckRoute(user, &team)
	if !b {
		utility.ResponseError(c, "管理员权限不足")
		return
	}

	if team.Status != 5 {
		utility.ResponseError(c, "请先扫团队扫码")
		return
	}

	if person.WalkStatus == 4 || person.WalkStatus == 5 {
		utility.ResponseError(c, "成员已结束毅行")
		return
	}

	if postForm.WalkStatus == 1 {
		person.WalkStatus = 3
	} else if postForm.WalkStatus == 2 {
		person.WalkStatus = 4
	} else {
		utility.ResponseError(c, "参数错误")
		return
	}
	userService.Update(*person)

	utility.ResponseSuccess(c, nil)
}

type UserSDForm struct {
	UserID     string `json:"user_id" binding:"required"`
	WalkStatus uint   `json:"walk_status" binding:"required"`
}

// UserSD 手动输入userID
func UserSD(c *gin.Context) {
	var postForm UserSDForm
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	user, err := adminService.GetAdminByJWT(c)

	// 获取个人信息
	person, err := model.GetPerson(postForm.UserID)
	if err != nil {
		utility.ResponseError(c, "用户ID错误")
		return
	}

	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)

	b := middleware.CheckRoute(user, &team)
	if !b {
		utility.ResponseError(c, "管理员权限不足")
		return
	}

	if team.Status != 5 {
		utility.ResponseError(c, "请先扫团队扫码")
		return
	}

	if person.WalkStatus == 4 || person.WalkStatus == 5 {
		utility.ResponseError(c, "成员已结束毅行")
		return
	}

	if postForm.WalkStatus == 1 {
		person.WalkStatus = 3
	} else if postForm.WalkStatus == 2 {
		person.WalkStatus = 4
	} else {
		utility.ResponseError(c, "参数错误")
		return
	}
	userService.Update(*person)

	utility.ResponseSuccess(c, nil)
}
