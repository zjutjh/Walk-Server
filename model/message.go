package model

import (
	"errors"
	"walk-server/global"
)

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

func InsertMessages(messages *[]Message) {
	global.DB.Create(messages)
}

func GetMessages(receiverOpenID string) ([]Message, error) {
	var messages []Message
	result := global.DB.Where("receiver_open_id = ?", receiverOpenID).Find(&messages)
	if result.RowsAffected == 0 {
		return nil, errors.New("no result")
	} else {
		return messages, nil	
	}
}

func DeleteMessage(id uint) {
	global.DB.Delete(&Message{}, id)
}