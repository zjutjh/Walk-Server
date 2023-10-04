package initial

import (
	"log"
	"walk-server/constant"
	"walk-server/global"
)

func ConstantInit() {
	if !global.Config.IsSet("number.ZH") || !global.Config.IsSet("number.PF_Half") || !global.Config.IsSet("number.PF_All") || !global.Config.IsSet("number.MGS_Half") || !global.Config.IsSet("number.MGS_All") {
		log.Fatal("点位数量未设置")
	}
	constant.PointMap[1] = uint8(global.Config.GetInt("number.ZH"))
	constant.PointMap[2] = uint8(global.Config.GetInt("number.PF_Half"))
	constant.PointMap[3] = uint8(global.Config.GetInt("number.PF_All"))
	constant.PointMap[4] = uint8(global.Config.GetInt("number.MGS_Half"))
	constant.PointMap[5] = uint8(global.Config.GetInt("number.MGS_All"))
}
