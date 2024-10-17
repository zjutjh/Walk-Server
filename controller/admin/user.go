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

type UserStatusList struct {
	List []UserStatusForm `json:"list" binding:"required"`
}

// UserSU 输入userID
func UserStatus(c *gin.Context) {
	var postForm UserStatusList
	err := c.ShouldBindJSON(&postForm)
	if err != nil {
		utility.ResponseError(c, "参数错误")
		return
	}

	user, _ := adminService.GetAdminByJWT(c)
	for _, form := range postForm.List {
		// 获取个人信息
		person, err := model.GetPerson(form.UserID)
		if err != nil {
			utility.ResponseError(c, "扫码错误，查找用户失败，请再次核对")
			return
		}

		var team model.Team
		if err := global.DB.Where("id = ?", person.TeamId).Take(&team).Error; err != nil {
			utility.ResponseError(c, "队伍信息获取失败")
			return
		}

		// 管理员只能管理自己所在的校区
		if !middleware.CheckRoute(user, &team) {
			utility.ResponseError(c, "该队伍为其他路线")
			return
		}

		if person.WalkStatus == 5 {
			utility.ResponseError(c, "成员已结束毅行")
			return
		}

		if form.Status == 1 {
			person.WalkStatus = 3
		} else {
			person.WalkStatus = 4
		}
		userService.Update(*person)
	}

	utility.ResponseSuccess(c, nil)
}
