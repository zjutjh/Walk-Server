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

type ModifyInfoRequest struct {
	Campus  uint8  `json:"campus" binding:"required"`
	College string `json:"college" binding:"required"`
	ID      string `json:"id" binding:"required"`
	Contact struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel" binding:"required"`
	} `json:"contact" binding:"required"`
}

func ModifyInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		var req ModifyInfoRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			reply.Fail(c, comm.CodeParameterInvalid)
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

		person.Campus = req.Campus
		person.College = req.College
		person.Identity = req.ID
		person.Qq = req.Contact.QQ
		person.Wechat = req.Contact.Wechat
		person.Tel = req.Contact.Tel

		if err := db.Save(&person).Error; err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
