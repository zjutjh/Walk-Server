package message

import (
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// ListMessage 获取自己应该接收的邮件
func ListMessage(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	messages, err := model.GetMessages(jwtData.OpenID)
	if err != nil {
		utility.ResponseError(context, "no result")
	}

	var messageRespData []gin.H
	for _, message := range messages {
		messageRespData = append(messageRespData, gin.H{
			"id": message.ID,
			"sender_open_id": message.SenderOpenId,
			"receiver_open_id": message.ReceiverOpenId,
			"message": message.Message,
		})
	}
	

	utility.ResponseSuccess(context, gin.H{
		"messages": messageRespData,
	})
}
