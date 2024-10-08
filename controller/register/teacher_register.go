package register

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// TeacherRegisterData 定义接收校友报名用的数据的类型
type TeacherRegisterData struct {
	Name    string `json:"name" binding:"required"`
	ID      string `json:"id" binding:"required"`
	StuID   string `json:"stu_id" binding:"required"`
	Gender  int8   `json:"gender" binding:"required"`
	Contact struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel" binding:"required"`
	} `json:"contact" binding:"required"`
}

func TeacherRegister(context *gin.Context) {
	// 获取 openID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken) // 中间件校验过是否合法了

	// 获取报名数据
	var postData TeacherRegisterData
	err := context.ShouldBindJSON(&postData)
	if err != nil {
		utility.ResponseError(context, "上传数据错误")
		return
	}

	person := model.Person{
		OpenId:     jwtData.OpenID,
		StuId:      postData.StuID,
		Name:       postData.Name,
		Gender:     postData.Gender,
		Campus:     5,
		Identity:   postData.ID,
		Status:     0,
		Qq:         postData.Contact.QQ,
		Wechat:     postData.Contact.Wechat,
		Tel:        postData.Contact.Tel,
		CreatedOp:  2,
		JoinOp:     5,
		TeamId:     -1,
		WalkStatus: 1,
		Type:       2,
	}

	result := global.DB.Create(&person)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "报名失败，请重试")
	} else {
		utility.ResponseSuccess(context, nil)
	}
}
