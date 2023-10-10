package teamService

import (
	"walk-server/global"
	"walk-server/model"
)

func Update(a model.Team) {
	global.DB.Save(&a)
}
