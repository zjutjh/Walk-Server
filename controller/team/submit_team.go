package team

import (
	"fmt"
	"strconv"
	"walk-server/global"
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 编写Lua脚本 - 判断是否已经提交、判断是否达到上限、提交后将剩余数量减一并记录提交的团队id
var submit = redis.NewScript(`
local teamID = KEYS[1];
local dailyRouteKey = KEYS[2];

local teamExists = redis.call("sismember", "teams", teamID);
if tonumber(teamExists) == 1 then
	return 1;
end

local num=redis.call("get", dailyRouteKey);
if tonumber(num) <= 0 then
	return 2;
end

redis.call("decr", dailyRouteKey);
redis.call("SAdd", "teams", teamID);
return 0;
`)

func SubmitTeam(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 查找用户
	person, _ := model.GetPerson(jwtData.OpenID)

	// 判断用户权限
	if person.Status == 0 {
		utility.ResponseError(context, "请先加入队伍")
		return
	} else if person.Status == 1 {
		utility.ResponseError(context, "没有修改的权限")
		return
	}

	var team model.Team

	result := global.DB.Where("id = ?", person.TeamId).Take(&team)
	if result.Error != nil {
		utility.ResponseError(context, "系统异常，请重试")
		return
	} else if team.Num < 4 {
		utility.ResponseError(context, "队伍人数不足四人")
		return
	}

	teamID := strconv.Itoa(int(team.ID))
	dailyRoute := utility.GetCurrentDate()*10 + team.Route
	dailyRouteKey := strconv.Itoa(int(dailyRoute))
	// 运行Lua脚本
	n, err := submit.Run(global.Rctx, global.Rdb, []string{teamID, dailyRouteKey}).Result()
	fmt.Println(err)
	if err != nil {
		utility.ResponseError(context, "系统异常，请重试")
		return
	}

	if n.(int64) == 1 {
		utility.ResponseError(context, "队伍已提交")
		return
	} else if n.(int64) == 2 {
		utility.ResponseError(context, "队伍数量已经到达上限，无法提交")
		return
	}

	utility.ResponseSuccess(context, nil)
}
