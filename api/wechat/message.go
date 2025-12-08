package wechat

import (
	"app/comm"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel/messages"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/zjutjh/mygo/foundation/reply"
)

// SendMessageHandler sends a message to a user
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

		oa := do.MustInvoke[*officialAccount.OfficialAccount](nil)
		msg := messages.NewText(req.Content)
		messenger := oa.CustomerService.Message(c.Request.Context(), msg)
		messenger.To = req.OpenID
		_, err := messenger.Send(c.Request.Context())
		if err != nil {
			reply.Fail(c, comm.CodeThirdServiceError)
			return
		}

		reply.Success(c, nil)
	}
}
