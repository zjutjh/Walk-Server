package model

type Admin struct {
	ID           uint   `json:"admin_id"`
	WechatOpenID string `json:"-"`
	Name         string `json:"name"`
	Account      string `json:"account"`
	Password     string `json:"-"`
	Point        uint8  `json:"point"`
	Route        uint8  `json:"route"` // 1 是朝晖路线，2 屏峰半程，3 屏峰全程，4 莫干山半程，5 莫干山全程
}
