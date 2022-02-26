package team

import (
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

type GetRandomListData struct {
	Route int `json:"route" binding:"required"`
}

func addTeamData(teamList []gin.H, teamResultSet *[]model.Team) []gin.H {
	for _, team := range *teamResultSet {
		teamList = append(teamList, gin.H{
			"id":     team.ID,
			"name":   team.Name,
			"num":    team.Num,
			"slogan": team.Slogan,
		})
	}

	return teamList
}

func GetRandomList(context *gin.Context) {
	// 解析请求数据
	var getRandomListData GetRandomListData
	err := context.ShouldBindJSON(&getRandomListData)
	if err != nil { // 参数发送错误
		utility.ResponseError(context, "参数错误")
		return
	}

	// 获取列表
	var teams []model.Team
	var teamList []gin.H

	// 先查找 3 人以下的团队
	result := initial.DB.Where("num <= 3 and allow_match = 1 and route = ?", getRandomListData.Route).Limit(3).Find(&teams)
	teamNum1 := result.RowsAffected
	teamList = addTeamData(teamList, &teams)

	// 查找 4 人团队
	result = initial.DB.Where("num = 4 and allow_match = 1 and route = ?", getRandomListData.Route).Limit(1 + (3 - int(teamNum1))).Find(&teams)
	teamNum2 := result.RowsAffected
	teamList = addTeamData(teamList, &teams)

	// 查找 5 人团队
	result = initial.DB.Where("num = 5 and allow_match = 1 and route = ?", getRandomListData.Route).Limit(1 + (1 - int(teamNum2)) + (3 - int(teamNum1))).Find(&teams)
	teamNum3 := result.RowsAffected
	teamList = addTeamData(teamList, &teams)

	if teamNum1+teamNum2+teamNum3 == 0 { // 没有查询结果
		utility.ResponseError(context, "No result")
	} else {
		utility.ResponseSuccess(context, gin.H{
			"teams": teamList,
		})
	}
}
