package register

import (
	"app/comm"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
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

		personRepo := repo.NewPersonRepo()
		person, err := personRepo.FindByStuId(c.Request.Context(), req.StuID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if person != nil {
			reply.Fail(c, comm.WithMsg(comm.CodeDataConflict, comm.MsgStudentAlreadyRegistered))
			return
		}

		openID := c.GetString("uid")
		if openID == "" {
			reply.Fail(c, comm.CodeNotLoggedIn)
			return
		}

		// 创建新用户
		newPerson := model.Person{
			StuId:    req.StuID,
			Name:     req.StuID, // Name 暂用学号代替，后续可能需要获取或传递
			Identity: req.ID,
			Campus:   req.Campus,
			College:  req.College,
			Type:     comm.PersonTypeStudent,
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
