package utility

import (
	"errors"
	"walk-server/global"
	"walk-server/model"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
)

// SendMessageToMembers 队长将消息发给队员
func SendMessageToMembers(msg string, captain model.Person, members []model.Person) {
	var messages []model.Message

	for _, member := range members {
		messages = append(messages, model.Message{
			SenderOpenId:   captain.OpenId,
			ReceiverOpenId: member.OpenId,
			Message:        msg,
		})

		SendMessageWithWechat(msg, member.OpenId)
	}

	model.InsertMessages(&messages)
}

// SendMessageToTeam 系统发送消息给所有的队员
func SendMessageToTeam(msg string, captain model.Person, members []model.Person) {
	var messages []model.Message

	// 添加发给队长的消息
	messages = append(messages, model.Message{
		Message:        msg,
		SenderOpenId:   "",
		ReceiverOpenId: captain.OpenId,
	})
	SendMessageWithWechat(msg, captain.OpenId)

	for _, member := range members {
		messages = append(messages, model.Message{
			Message:        msg,
			SenderOpenId:   "",
			ReceiverOpenId: member.OpenId,
		})

		SendMessageWithWechat(msg, member.OpenId)
	}

	model.InsertMessages(&messages)
}

// SendMessage 人和人发送消息
func SendMessage(msg string, sender *model.Person, receiver *model.Person) {
	if sender == nil { // 系统消息
		model.InsertMessage(msg, "", receiver.OpenId)
	} else {
		model.InsertMessage(msg, sender.OpenId, receiver.OpenId)
	}

	SendMessageWithWechat(msg, receiver.OpenId)
}

func DeleteMessage(id uint, jwtData *JwtData) error {
	// 校验是否是接收者删除了自己的消息
	var message model.Message
	global.DB.Where("id = ?", id).Take(&message)
	if message.ReceiverOpenId != jwtData.OpenID {
		return errors.New("access denied")
	}

	model.DeleteMessage(id)
	return nil
}

func SendMessageWithWechat(msg string, receiverEncOpenID string) {
	// 解密 open ID
	aesKey := global.Config.GetString("server.AESSecret")
	receiverOpenID := AesDecrypt(receiverEncOpenID, aesKey)

	text := messages.NewText(msg + "\n---\n因为微信的限制，请回复'收到'以确保后续消息的正常接收")
	_, err := global.OfficialAccount.CustomerService.Message(global.Wctx, text).SetTo(receiverOpenID).Send(global.Wctx)

	if err != nil {
		// log.Println(err) // Optional logging
	}
}
