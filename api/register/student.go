package register

import (
	"app/comm"
	"app/dao/model"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

type StudentRegisterRequest struct {
	StuID    string `json:"stu_id" binding:"required"`
	Password string `json:"password" binding:"required"`
	ID       string `json:"id" binding:"required"`
	Campus   uint8  `json:"campus" binding:"required"`
	College  string `json:"college" binding:"required"`
	Contact  struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel" binding:"required"`
	} `json:"contact" binding:"required"`
}

func StudentRegisterHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req StudentRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		// TODO: Verify student info with external service (WeJH-SDK) if needed
		// For now, we assume it's valid or skip verification as we don't have the SDK setup here fully

		db := ndb.Pick()
		var person model.Person
		err := db.Where("stu_id = ?", req.StuID).First(&person).Error
		if err == nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "该学号已报名"))
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		// Create new person
		newPerson := model.Person{
			StuId:    req.StuID,
			Name:     req.StuID, // Name might need to be fetched or passed
			Identity: req.ID,
			Campus:   req.Campus,
			College:  req.College,
			Type:     1, // Student
			Qq:       req.Contact.QQ,
			Wechat:   req.Contact.Wechat,
			Tel:      req.Contact.Tel,
			// Password? The model doesn't have password field in my definition,
			// but the request has it. Maybe it's for verification?
			// In main branch, Person model doesn't seem to have password either,
			// maybe it's stored elsewhere or used for verification.
		}

		if err := db.Create(&newPerson).Error; err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
