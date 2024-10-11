package userService

import (
	"walk-server/global"
	"walk-server/model"
)

func GetUserByID(id string) (*model.Person, error) {
	var person model.Person
	result := global.DB.Where("identity = ?", id).First(&person)
	return &person, result.Error
}
