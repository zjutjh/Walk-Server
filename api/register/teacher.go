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

type TeacherRegisterRequest struct {
	ID       string `json:"id" binding:"required"`
	StuID    string `json:"stu_id" binding:"required"` // 工号?
	Password string `json:"password" binding:"required"`
	Campus   uint8  `json:"campus"`
	Contact  struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel" binding:"required"`
	} `json:"contact" binding:"required"`
}

func TeacherRegisterHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TeacherRegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		db := ndb.Pick()
		var person model.Person
		err := db.Where("stu_id = ?", req.StuID).First(&person).Error
		if err == nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, "该工号已报名"))
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		newPerson := model.Person{
			StuId:    req.StuID,
			Identity: req.ID,
			Campus:   req.Campus,
			Type:     2, // Teacher
			Qq:       req.Contact.QQ,
			Wechat:   req.Contact.Wechat,
			Tel:      req.Contact.Tel,
		}

		if err := db.Create(&newPerson).Error; err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
