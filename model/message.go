package model

type Message struct {
	ID             uint
	SenderOpenId   string
	ReceiverOpenId string `gorm:"index"`
	Message        string
}
