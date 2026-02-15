package user

import (
	"app/comm"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
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

		personRepo := repo.NewPersonRepo()
		person, err := personRepo.FindByOpenId(c.Request.Context(), openID)
		if err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}
		if person == nil {
			reply.Fail(c, comm.CodeDataNotFound)
			return
		}

		person.Campus = req.Campus
		person.College = req.College
		person.Identity = req.ID
		qq := req.Contact.QQ
		wechat := req.Contact.Wechat
		person.QQ = &qq
		person.Wechat = &wechat
		person.Tel = req.Contact.Tel

		// 更新用户
		if err := personRepo.Update(c.Request.Context(), nil, person); err != nil {
			reply.Fail(c, comm.CodeDatabaseError)
			return
		}

		reply.Success(c, nil)
	}
}
