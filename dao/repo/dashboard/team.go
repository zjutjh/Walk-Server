package repo

import (
	"context"
	"time"

	"app/dao/model"
)

// GetTeamByID 查询指定队伍。
func (r *DashboardRepo) GetTeamByID(ctx context.Context, teamID int64) (*model.Team, error) {
	return r.query.Team.WithContext(ctx).
		Where(r.query.Team.ID.Eq(teamID)).
		First()
}

// UpdateTeamLostStatus 更新队伍失联状态。仅当 isLost=false 时更新时间戳。
func (r *DashboardRepo) UpdateTeamLostStatus(ctx context.Context, teamID int64, isLost bool, statusUpdatedAt time.Time) (updated bool, err error) {
	isLostVal := int8(0)
	if isLost {
		isLostVal = 1
	}

	m := map[string]interface{}{
		"is_lost": isLostVal,
	}

	// 仅当 isLost=false 时更新时间戳
	if !isLost {
		m["time"] = statusUpdatedAt
	}

	tx := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams").
		Where("id = ?", teamID).
		Updates(m)
	if tx.Error != nil {
		return false, tx.Error
	}

	return tx.RowsAffected > 0, nil
}
