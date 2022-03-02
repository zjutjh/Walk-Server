package utility

import "walk-server/model"

// TODO 加上微信通知功能

// SaveMessageToMembers 队长将消息发给队员
func SendMessageToMembers(message string, captain model.Person, members []model.Person) {
	for _, member := range members {
		model.InsertMessage(message, captain.OpenId, member.OpenId)
	}
}

// SendMessageToTeam 系统发送消息给所有的队员
func SendMessageToTeam(message string, persons []model.Person) {
	for _, person := range persons {
		model.InsertMessage(message, "", person.OpenId) // 发送人 open ID 为空表示这是系统消息
	}
}

// SendMessage 人和人发送消息
func SendMessage(message string, sender model.Person, receiver model.Person) {
	model.InsertMessage(message, sender.OpenId, receiver.OpenId)
}