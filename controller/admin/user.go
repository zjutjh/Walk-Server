package admin

import (
	"walk-server/global"
	"walk-server/middleware"
	"walk-server/model"
	"walk-server/service/adminService"
	"walk-server/service/userService"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

type UserStatusForm struct {
	UserID string `json:"user_id" binding:"required"`
	Status int    `json:"status" binding:"required,oneof=1 2"`
}

// UserSU 输入userID
func UserStatus(c *gin.Context) {
	var postForm UserStatusForm
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	user, _ := adminService.GetAdminByJWT(c)

	// 获取个人信息

	person, err := model.GetPerson(postForm.UserID)
	if err != nil {
		utility.ResponseError(c, "用户ID错误")
		return
	}

	var team model.Team
	global.DB.Where("id = ?", person.TeamId).Take(&team)
	// 管理员只能管理自己所在的校区
	b := middleware.CheckRoute(user, &team)
	if !b {
		utility.ResponseError(c, "管理员权限不足")
		return
	}

	if team.Status != 5 {
		utility.ResponseError(c, "请先扫团队扫码")
		return
	}

	if person.WalkStatus == 5 {
		utility.ResponseError(c, "成员已结束毅行")
		return
	}

	if postForm.Status == 1 {
		person.WalkStatus = 3
	} else {
		person.WalkStatus = 4
	}
	userService.Update(*person)

	utility.ResponseSuccess(c, nil)
}
