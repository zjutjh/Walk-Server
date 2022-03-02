package model

import "walk-server/global"

type Message struct {
	ID             uint
	SenderOpenId   string // 如果发送者 open ID 为空, 相当于系统消息
	ReceiverOpenId string `gorm:"index"`
	Message        string
}

func InsertMessage(message string, senderOpenID string, receiverOpenID string) {
	global.DB.Create(&Message{
		SenderOpenId: senderOpenID,
		ReceiverOpenId: receiverOpenID,
		Message: message,
	})
}
