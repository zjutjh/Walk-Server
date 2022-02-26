package team

import (
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func addTeamData(teamList []gin.H, teamResultSet *[]model.Team) []gin.H {
	for _, team := range *teamResultSet {
		teamList = append(teamList, gin.H{
			"name":   team.Name,
			"num":    team.Num,
			"slogan": team.Slogan,
		})
	}

	return teamList
}

func GetRandomList(context *gin.Context) {
	var teams []model.Team
	var teamList []gin.H

	// 先查找 3 人以下的团队
	result := initial.DB.Where("num <= 3").Limit(3).Find(&teams)
	teamNum1 := result.RowsAffected
	teamList = addTeamData(teamList, &teams)

	// 查找 4 人团队
	result = initial.DB.Where("num = 4").Limit(1 + (3 - int(teamNum1))).Find(&teams)
	teamNum2 := result.RowsAffected
	teamList = addTeamData(teamList, &teams)

	// 查找 5 人团队
	initial.DB.Where("num = 5").Limit(1 + (1 - int(teamNum2)) + (3 - int(teamNum1))).Find(&teams)
	teamList = addTeamData(teamList, &teams)

	utility.ResponseSuccess(context, gin.H{
		"teams": teamList,
	})
}
