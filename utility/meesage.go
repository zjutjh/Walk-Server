package utility

import (
	"errors"
	"fmt"
	"walk-server/global"
	"walk-server/model"

	"github.com/go-resty/resty/v2"
)

// SaveMessageToMembers 队长将消息发给队员
func SendMessageToMembers(message string, captain model.Person, members []model.Person) {
	var messages []model.Message

	for _, member := range members {
		messages = append(messages, model.Message{
			SenderOpenId:   captain.OpenId,
			ReceiverOpenId: member.OpenId,
			Message:        message,
		})

		SendMessageWithWechat(message, member.OpenId)
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
	SendMessageWithWechat(message, captain.OpenId)

	for _, member := range members {
		messages = append(messages, model.Message{
			Message:        message,
			SenderOpenId:   "",
			ReceiverOpenId: member.OpenId,
		})

		SendMessageWithWechat(message, member.OpenId)
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

	SendMessageWithWechat(message, receiver.OpenId)
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

func SendMessageWithWechat(message string, receiverEncOpenID string) {
	var accessToekn string
	var err error
	// accessToekn = global.Config.GetString("server.accessToken")
	wechatAPPID := global.Config.GetString("server.wechatAPPID")
	wechatSecret := global.Config.GetString("server.wechatSecret")
	accessToekn, err = GetAccessToken(wechatAPPID, wechatSecret)

	if err != nil {
		if IsDebugMode() {
			fmt.Println(err)
		}
		return
	}

	// 解密 open ID
	aesKey := global.Config.GetString("server.AESSecret")
	receiverOpenID := AesDecrypt(receiverEncOpenID, aesKey)

	client := resty.New()
	resp, err := client.R().SetBody(map[string]interface{}{
		"touser":  receiverOpenID,
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": message + "\n---\n因为微信的限制，请回复'收到'以确保后续消息的正常接收",
		},
	}).Post("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=" + accessToekn)

	if IsDebugMode() {
		fmt.Println(string(resp.Body()))
		fmt.Println(err)
	}
}
