package wechat

import (
	"app/comm"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	oa "github.com/zjutjh/mygo/wechat/officialAccount"
)

// SendMessageHandler 发送模板消息给用户
func SendMessageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OpenID  string `json:"open_id"`
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		officialAccount := oa.Pick()
		msg := messages.NewText(req.Content)
		messenger := officialAccount.CustomerService.Message(c.Request.Context(), msg)
		messenger.To = req.OpenID
		_, err := messenger.Send(c.Request.Context())
		if err != nil {
			reply.Fail(c, comm.CodeThirdServiceError)
			return
		}

		reply.Success(c, nil)
	}
}
