package register

import (
	"app/comm"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
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

		personRepo := repo.NewPersonRepo()
		person, err := personRepo.FindByStuId(c.Request.Context(), req.StuID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if person != nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, comm.MsgTeacherAlreadyRegistered))
			return
		}

		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		newPerson := model.Person{
			StuId:    req.StuID,
			Identity: req.ID,
			Campus:   req.Campus,
			Type:     comm.PersonTypeTeacher,
			Qq:       req.Contact.QQ,
			Wechat:   req.Contact.Wechat,
			Tel:      req.Contact.Tel,
			OpenId:   openID,
		}

		if err := personRepo.Create(c.Request.Context(), nil, &newPerson); err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
