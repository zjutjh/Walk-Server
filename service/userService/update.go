package userService

import (
	"walk-server/model"
)

func Update(a model.Person) {
	model.UpdatePerson(a.OpenId, &a)
}
