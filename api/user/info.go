package user

import (
	"app/comm"
	"app/dao/model"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

func GetInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		db := ndb.Pick()
		var person model.Person
		err := db.Where("open_id = ?", openID).First(&person).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, person)
	}
}
