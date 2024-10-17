package userService

import (
	"walk-server/model"
)

func Update(a model.Person) {
	model.UpdatePerson(a.OpenId, &a)
}

func Set(open_id string, a model.Person) error {
	return model.SetPerson(open_id, &a)
}
