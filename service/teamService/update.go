package teamService

import (
	"walk-server/global"
	"walk-server/model"
)

func Update(a model.Team) {
	global.DB.Save(&a)
}

func Delete(a model.Team) error {
	return global.DB.Delete(&a).Error
}

func Create(a model.Team) {
	global.DB.Create(&a)
}

func UpdateCaptain(teamID int, openID string) error {
	return global.DB.Exec("UPDATE teams SET captain = ? WHERE id = ?", openID, teamID).Error
}
