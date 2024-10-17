package register

import (
	"github.com/gin-gonic/gin"
	"walk-server/service/userService"
	"walk-server/utility"
)

type LoginData struct {
	Name string `json:"name" binding:"required"`
	ID   string `json:"id" binding:"required"`
	Tel  string `json:"tel" binding:"required"`
}

func Login(context *gin.Context) {
	// 获取 openID
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken) // 中间件校验过是否合法了

	var postForm LoginData
	err := context.ShouldBindJSON(&postForm)

	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}
	user, err := userService.GetUserByID(postForm.ID)
	if err != nil {
		utility.ResponseError(context, "信息错误,请检查是否填写有误")
		return
	}
	if user.Tel != postForm.Tel && user.Name != postForm.Name {
		utility.ResponseError(context, "信息错误,请检查是否填写有误")
		return
	}
	user.OpenId = jwtData.OpenID
	userService.Update(*user)
	utility.ResponseSuccess(context, nil)
}
