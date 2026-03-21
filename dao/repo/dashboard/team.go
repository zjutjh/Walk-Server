package repo

import (
	"context"
	"sort"
	"strings"
	"time"

	"app/dao/model"
)

// TeamMemberRow 队伍成员信息行。
type TeamMemberRow struct {
	ID     int64
	OpenID string
	Name   string
	Phone  string
	Role   string
}

// GetTeamByID 查询指定队伍。
func (r *DashboardRepo) GetTeamByID(ctx context.Context, teamID int64) (*model.Team, error) {
	return r.query.Team.WithContext(ctx).
		Where(r.query.Team.ID.Eq(teamID)).
		First()
}

// ListTeamMembers 查询队伍内成员，按队长优先、ID 升序返回。
func (r *DashboardRepo) ListTeamMembers(ctx context.Context, teamID int64) ([]TeamMemberRow, error) {
	peopleRows, err := r.query.People.WithContext(ctx).
		Where(r.query.People.TeamID.Eq(teamID)).
		Find()
	if err != nil {
		return nil, err
	}

	members := make([]TeamMemberRow, 0, len(peopleRows))
	for _, row := range peopleRows {
		members = append(members, TeamMemberRow{
			ID:     row.ID,
			OpenID: row.OpenID,
			Name:   row.Name,
			Phone:  row.Tel,
			Role:   row.Role,
		})
	}

	sort.Slice(members, func(i, j int) bool {
		leftRank := roleSortRank(members[i].Role)
		rightRank := roleSortRank(members[j].Role)
		if leftRank != rightRank {
			return leftRank < rightRank
		}

		return members[i].ID < members[j].ID
	})

	return members, nil
}

// roleSortRank 统一角色排序：队长优先，其余角色其后。
func roleSortRank(role string) int {
	if strings.EqualFold(role, "captain") {
		return 0
	}

	return 1
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
