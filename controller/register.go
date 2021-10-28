package controller

import (
	"github.com/gin-gonic/gin"
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"
)

// StudentRegisterData 定义接收学生报名用的数据的类型
type StudentRegisterData struct {
	Name    string `json:"name"`
	StuID   string `json:"stu_id"`
	ID      string `json:"id"`
	Gender  uint8  `json:"gender"`
	Campus  uint8  `json:"campus"`
	Contact struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel"`
	} `json:"contact"`
}

// GraduateRegisterData 定义接收校友报名用的数据的类型
type GraduateRegisterData struct {
	Name    string `json:"name" binding:"required"`
	ID      string `json:"id" binding:"required"`
	Gender  uint8  `json:"gender" binding:"required"`
	Contact struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel"`
	} `json:"contact"`
	Healthcodeimgurl string `json:"healthcodeimgurl" binding:"required"`
}

func StudentRegister(context *gin.Context) {
	// 获取 openID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken) // 中间件校验过是否合法了

	// 获取报名数据
	var postData StudentRegisterData
	err := context.ShouldBindJSON(&postData)
	if err != nil {
		utility.ResponseError(context, "上传数据错误")
		return
	}

	person := model.Person{
		OpenId:    jwtData.OpenID,
		Name:      postData.Name,
		Gender:    postData.Gender,
		StuId:     postData.StuID,
		Campus:    postData.Campus,
		Qq:        postData.Contact.QQ,
		Wechat:    postData.Contact.Wechat,
		Tel:       postData.Contact.Tel,
		CreatedOp: 1,
		JoinOp:    3,
		TeamId:    "",
	}

	result := initial.DB.Create(&person)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "报名失败，请重试")
	} else {
		utility.ResponseSuccess(context, nil)
	}
}

func GraduateRegister(context *gin.Context) {
	// 获取 openID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken) // 中间件校验过是否合法了

	// 获取报名数据
	var postData GraduateRegisterData
	err := context.ShouldBindJSON(&postData)
	if err != nil {
		utility.ResponseError(context, "上传数据错误")
		return
	}

	person := model.Person{
		OpenId:    jwtData.OpenID,
		Name:      postData.Name,
		Gender:    postData.Gender,
		Campus:    5,
		Qq:        postData.Contact.QQ,
		Wechat:    postData.Contact.Wechat,
		Tel:       postData.Contact.Tel,
		CreatedOp: 1,
		JoinOp:    3,
		TeamId:    "",
	}

	result := initial.DB.Create(&person)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "报名失败，请重试")
	} else {
		utility.ResponseSuccess(context, nil)
	}
}
