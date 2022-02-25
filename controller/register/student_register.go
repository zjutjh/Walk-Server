package register

import (
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

// StudentRegisterData 定义接收学生报名用的数据的类型
type StudentRegisterData struct {
	Name    string `json:"name"`
	StuID   string `json:"stu_id"`
	ID      string `json:"id"`
	Gender  int8   `json:"gender"`
	College string `json:"college"`
	Campus  uint8  `json:"campus"`
	Contact struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel"`
	} `json:"contact"`
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
		Status:    0,
		College:   postData.College,
		Identity:  postData.ID,
		Campus:    postData.Campus,
		Qq:        postData.Contact.QQ,
		Wechat:    postData.Contact.Wechat,
		Tel:       postData.Contact.Tel,
		CreatedOp: 2,
		JoinOp:    5,
		TeamId:    -1,
	}

	result := initial.DB.Create(&person)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "报名失败，请重试")
	} else {
		utility.ResponseSuccess(context, nil)
	}
}
