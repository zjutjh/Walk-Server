package repo

import (
	"context"
	"time"

	"github.com/zjutjh/mygo/ndb"
)

type Message struct {
	ID         int64     `gorm:"column:id"`
	SenderID   *int64    `gorm:"column:sender_id"`
	ReceiverID int64     `gorm:"column:receiver_id"`
	Message    string    `gorm:"column:message"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (*Message) TableName() string {
	return "messages"
}

type MessageRepo struct{}

func NewMessageRepo() *MessageRepo {
	return &MessageRepo{}
}

func (r *MessageRepo) ListByReceiverID(ctx context.Context, receiverID int64) ([]Message, error) {
	messages := make([]Message, 0)
	err := ndb.Pick().WithContext(ctx).
		Where("receiver_id = ?", receiverID).
		Order("created_at DESC, id DESC").
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepo) DeleteByIDAndReceiverID(ctx context.Context, messageID, receiverID int64) error {
	return ndb.Pick().WithContext(ctx).
		Where("id = ? AND receiver_id = ?", messageID, receiverID).
		Delete(&Message{}).
		Error
}
