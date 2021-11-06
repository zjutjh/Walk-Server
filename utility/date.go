package utility

import (
	"time"
	"walk-server/utility/initial"
)

// GetCurrentDate 获取当前的天数
func GetCurrentDate() uint8 {
	var timeLayout = "2006-01-02 15:04:05"
	times, _ := time.Parse(timeLayout, initial.Config.GetString("startDate"))
	timeUnix := times.Unix()
	nowTimeUnix := time.Now().Unix() - timeUnix
	return uint8(nowTimeUnix / 3600 / 24)
}

func CanOpenApi() bool {
	var timeLayout = "2006-01-02 15:04:05"
	startTimes, _ := time.Parse(timeLayout, initial.Config.GetString("startDate"))
	startTimeUnix := startTimes.Unix()
	if time.Now().Unix() <= startTimeUnix {
		return false
	}

	if time.Now().Hour() >= 8 {
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
