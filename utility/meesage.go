package utility

import "walk-server/model"

// TODO 加上微信通知功能

// SaveMessageToMembers 队长将消息发给队员
func SendMessageToMembers(message string, captain model.Person, members []model.Person) {
	var messages []model.Message

	for _, member := range members {
		messages = append(messages, model.Message{
			SenderOpenId:   captain.OpenId,
			ReceiverOpenId: member.OpenId,
			Message:        message,
		})
	}

	model.InsertMessages(&messages)
}

// SendMessageToTeam 系统发送消息给所有的队员
func SendMessageToTeam(message string, captain model.Person, members []model.Person) {
	var messages []model.Message

	// 添加发给队长的消息
	messages = append(messages, model.Message{
		Message:        message,
		SenderOpenId:   "",
		ReceiverOpenId: captain.OpenId,
	})

	for _, member := range members {
		messages = append(messages, model.Message{
			Message:        message,
			SenderOpenId:   "",
			ReceiverOpenId: member.OpenId,
		})
	}

	model.InsertMessages(&messages)
}

// SendMessage 人和人发送消息
func SendMessage(message string, sender *model.Person, receiver *model.Person) {
	if sender == nil { // 系统消息
		model.InsertMessage(message, "", receiver.OpenId)
	} else {
		model.InsertMessage(message, sender.OpenId, receiver.OpenId)
	}
}
