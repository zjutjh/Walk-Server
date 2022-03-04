package message

import (
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

type DeleteMessageData struct {
	ID uint `json:"message_id"`
}

func DeleteMessage(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 获取消息 ID
	var deleteMessageData DeleteMessageData
	err := context.ShouldBindJSON(&deleteMessageData)
	if err != nil {
		utility.ResponseError(context, "上传数据错误")
		return
	}

	err = utility.DeleteMessage(deleteMessageData.ID, jwtData)
	if err != nil {
		utility.ResponseError(context, "access denied")
		return
	}

	utility.ResponseSuccess(context, nil)
}
