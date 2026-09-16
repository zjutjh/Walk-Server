package repo

import (
	"context"
	"errors"
	"time"

	"app/comm"
	"app/dao/model"
	"app/dao/query"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
)

const unreadNoticeLimit = 50

type NoticeRepo struct {
	query *query.Query
}

func NewNoticeRepo() *NoticeRepo {
	return &NoticeRepo{query: query.Use(ndb.Pick())}
}

func NewNoticeRepoWithTx(tx *query.Query) *NoticeRepo {
	return &NoticeRepo{query: tx}
}

func (r *NoticeRepo) ListUnread(ctx context.Context, userID int64) ([]*model.Notice, error) {
	n := r.query.Notice
	return n.WithContext(ctx).
		Where(n.UserID.Eq(userID), n.ReadAt.IsNull()).
		Order(n.CreatedAt, n.ID).
		Limit(unreadNoticeLimit).
		Find()
}

func (r *NoticeRepo) Ack(ctx context.Context, userID, noticeID int64) error {
	n := r.query.Notice
	_, err := n.WithContext(ctx).
		Where(n.UserID.Eq(userID), n.ID.Eq(noticeID), n.ReadAt.IsNull()).
		Update(n.ReadAt, time.Now())
	return err
}

func (r *NoticeRepo) UpsertUnread(ctx context.Context, notices []*model.Notice) error {
	if len(notices) == 0 {
		return nil
	}
	n := r.query.Notice
	for _, notice := range notices {
		existing, err := n.WithContext(ctx).
			Where(n.UserID.Eq(notice.UserID), n.Type.Eq(string(notice.Type)), n.ReadAt.IsNull()).
			Order(n.ID.Desc()).
			First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := n.WithContext(ctx).Create(notice); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if _, err := n.WithContext(ctx).Where(n.ID.Eq(existing.ID)).Updates(map[string]any{
			"actor_id":   notice.ActorID,
			"actor_name": notice.ActorName,
			"team_id":    notice.TeamID,
			"created_at": time.Now(),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *NoticeRepo) DeleteUnreadTypes(ctx context.Context, userID int64, types ...comm.NoticeType) error {
	if len(types) == 0 {
		return nil
	}
	n := r.query.Notice
	values := make([]string, 0, len(types))
	for _, noticeType := range types {
		values = append(values, string(noticeType))
	}
	_, err := n.WithContext(ctx).
		Where(n.UserID.Eq(userID), n.Type.In(values...), n.ReadAt.IsNull()).
		Delete()
	return err
}
