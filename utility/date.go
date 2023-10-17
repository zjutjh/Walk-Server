package utility

import (
	"time"
	"walk-server/global"
)

const TimeLayout = "2006-01-02 15:04:05"

// GetCurrentDate 获取当前的天数
func GetCurrentDate() uint8 {
	times, _ := time.ParseInLocation(
		TimeLayout,
		global.Config.GetString("startDate"),
		time.Local,
	)
	timeUnix := times.Unix()
	nowTimeUnix := time.Now().Unix() - timeUnix
	return uint8(nowTimeUnix / 3600 / 24)
}

func CanOpenApi() bool {
	startTimes, _ := time.ParseInLocation(
		TimeLayout,
		global.Config.GetString("startDate"),
		time.Local,
	)
	startTimeUnix := startTimes.Unix()
	if time.Now().Unix() <= startTimeUnix {
		return false
	}

	if time.Now().Hour() >= 6 {
		return true
	} else {
		return false
	}
}

func CanSubmit() bool {
	if time.Now().Hour() >= 12 {
		return true
	} else {
		return false
	}
}
