package controller

import (
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

type UserModifyData struct {
	Name    string `json:"name" binding:"required"`
	StuID   string `json:"stu_id"`
	ID      string `json:"id" binding:"required"`
	Gender  int8   `json:"gender" binding:"required"`
	Campus  uint8  `json:"campus"`
	Contact struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel"`
	} `json:"contact"`
}

func GetInfo(context *gin.Context) {
	// 获取 open ID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken) // 中间件校验过数据了
	openID := jwtData.OpenID

	// 获取用户数据
	person := model.Person{}
	initial.DB.Where("open_id = ?", openID).First(&person)

	utility.ResponseSuccess(context, gin.H{
		"name":      person.Name,
		"stu_id":    person.StuId,
		"gender":    person.Gender,
		"id":        person.Identity,
		"campus":    person.Campus,
		"status":    person.Status, 
		"create_op": person.CreatedOp,
		"join_op":   person.JoinOp,
		"team_id":   person.TeamId,
		"contact": gin.H{
			"qq":     person.Qq,
			"wechat": person.Wechat,
			"tel":    person.Tel,
		},
	})
}

func ModifyInfo(context *gin.Context) {
	// 获取 open ID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken) // 中间件校验过数据了
	openID := jwtData.OpenID

	// 获取 post data
	var postData UserModifyData
	err := context.ShouldBindJSON(&postData)
	if err != nil {
		utility.ResponseError(context, "参数错误，请重试")
		return
	}

	// 更新数据
	var person model.Person
	initial.DB.Model(&person).Where("open_id = ?", openID).Updates(model.Person{
		Name:     postData.Name,
		Gender:   postData.Gender,
		StuId:    postData.StuID,
		Campus:   postData.Campus,
		Identity: postData.ID,
		Qq:       postData.Contact.QQ,
		Wechat:   postData.Contact.Wechat,
		Tel:      postData.Contact.Tel,
	})
	utility.ResponseSuccess(context, nil)
}
