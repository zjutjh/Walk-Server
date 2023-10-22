package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func RedisToMysql(c *gin.Context) {
	var teams []model.Team
	if err := global.DB.Find(&teams).Error; err != nil {
		utility.ResponseError(c, "服务异常，请重试")
		return
	}
	for _, team := range teams {
		if exists, _ := global.Rdb.SIsMember(global.Rctx, "teams", team.ID).Result(); exists {
			team.Submit = true
		} else {
			team.Submit = false
		}
		if err := global.DB.Select("submit").Updates(team).Error; err != nil {
			utility.ResponseError(c, "服务异常，请重试")
			return
		}
	}
	utility.ResponseSuccess(c, nil)
}
