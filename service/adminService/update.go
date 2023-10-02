package adminService

import (
	"walk-server/global"
	"walk-server/model"
)

func UpdateOpenID(admin model.Admin) {
	global.DB.Updates(&admin)
}
