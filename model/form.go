package model

import (
	"time"
	"walk-server/global"
)

type Form struct {
	ID      uint      `json:"id"`       //主键
	AdminID uint      `json:"admin_id"` //管理员ID
	Route   uint8     `json:"route"`    //路线  1 是朝晖路线，2 屏峰半程，3 屏峰全程，4 莫干山半程，5 莫干山全程
	Point   int8      `json:"point"`    //点位
	Data    []byte    `json:"data"`     //数据
	Time    time.Time `json:"time"`     //时间
}

func InsertForm(form Form) error {
	return global.DB.Create(&form).Error
}
