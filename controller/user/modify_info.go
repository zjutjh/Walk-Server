package user

import (
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

type UserModifyData struct {
	Name    string `json:"name" binding:"required"`
	StuID   string `json:"stu_id"`
	ID      string `json:"id" binding:"required"`
	Gender  int8   `json:"gender" binding:"required"`
	College string `json:"college"`
	Campus  uint8  `json:"campus"`
	Contact struct {
		QQ     string `json:"qq"`
		Wechat string `json:"wechat"`
		Tel    string `json:"tel"`
	} `json:"contact"`
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
	model.UpdatePerson(openID, &model.Person{
		Name:     postData.Name,
		Gender:   postData.Gender,
		StuId:    postData.StuID,
		Campus:   postData.Campus,
		College:  postData.College,
		Identity: postData.ID,
		Qq:       postData.Contact.QQ,
		Wechat:   postData.Contact.Wechat,
		Tel:      postData.Contact.Tel,
	})
	utility.ResponseSuccess(context, nil)
}
