package team

import (
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// CreateTeamData 接收创建团队信息的数据类型
type CreateTeamData struct {
	Name       string `json:"name" binding:"required"`
	Route      uint8  `json:"route" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Slogan     string `json:"slogan" binding:"required"`
	AllowMatch *bool  `json:"allow_match" binding:"required"`
}

func CreateTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 获取 post json 数据
	var createTeamData CreateTeamData
	err := context.ShouldBindJSON(&createTeamData)
	if err != nil {
		utility.ResponseError(context, "参数错误")
		return
	}

	if createTeamData.Route > 6 {
		utility.ResponseError(context, "参数错误")
		return
	}

	// 查询用户信息
	person, _ := model.GetPerson(jwtData.OpenID)

	if person.Status != 0 { // 现在已经加入了一个团队
		utility.ResponseError(context, "请先退出或解散原来的团队")
		return
	}

	if person.CreatedOp == 0 {
		utility.ResponseError(context, "无法创建团队了")
		return
	}

	team := model.Team{
		Name:       createTeamData.Name,
		Num:        1,
		AllowMatch: *createTeamData.AllowMatch,
		Password:   createTeamData.Password,
		Captain:    person.OpenId,
		Route:      createTeamData.Route,
		Slogan:     createTeamData.Slogan,
		Point:      0,
		StartNum:   0,
		Status:     1,
	}

	// 事务
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		// 创建团队
		if err := tx.Create(&team).Error; err != nil {
			return err
		}

		// 加入团队后对应的状态更新
		person.CreatedOp -= 1
		person.Status = 2
		person.TeamId = int(team.ID)

		if err := model.TxUpdatePerson(tx, person); err != nil {
			return err
		}

		// 返回 nil 提交事务
		return nil
	})
	if err != nil {
		utility.ResponseError(context, "服务异常，请重试")
		return
	}

	// 返回 team_id
	utility.ResponseSuccess(context, gin.H{
		"team_id": team.ID,
	})
}
