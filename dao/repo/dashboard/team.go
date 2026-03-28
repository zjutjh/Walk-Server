package repo

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

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

type TeamFilterQuery struct {
	Campus        string
	ToPointName   string
	PrevPointName string
	Key           string
	SearchType    string
	Limit         int
	Offset        int
}

type TeamFilterRow struct {
	TeamID        int64        `gorm:"column:team_id"`
	CaptainName   string       `gorm:"column:captain_name"`
	CaptainPhone  string       `gorm:"column:captain_phone"`
	PrevPointName string       `gorm:"column:prev_point_name"`
	PrevPointTime sql.NullTime `gorm:"column:prev_point_time"`
	RouteName     string       `gorm:"column:route_name"`
	IsLost        int8         `gorm:"column:is_lost"`
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

// CountTeamsByFilter 统计筛选条件下的队伍总数。
func (r *DashboardRepo) CountTeamsByFilter(ctx context.Context, query TeamFilterQuery) (int64, error) {
	var total int64
	err := r.buildTeamFilterBaseQuery(ctx, query).
		Distinct("t.id").
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

// ListTeamsByFilter 查询筛选条件下的队伍列表。
func (r *DashboardRepo) ListTeamsByFilter(ctx context.Context, query TeamFilterQuery) ([]TeamFilterRow, error) {
	rows := make([]TeamFilterRow, 0)

	err := r.buildTeamFilterBaseQuery(ctx, query).
		Select("t.id AS team_id, COALESCE(p.name, '') AS captain_name, COALESCE(p.tel, '') AS captain_phone, t.prev_point_name, t.time AS prev_point_time, t.route_name, t.is_lost").
		Order("t.time ASC").
		Order("t.id ASC").
		Limit(query.Limit).
		Offset(query.Offset).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (r *DashboardRepo) buildTeamFilterBaseQuery(ctx context.Context, query TeamFilterQuery) *gorm.DB {
	db := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams AS t").
		Joins("JOIN routes AS r ON r.name = t.route_name AND r.is_active = ? AND r.campus = ?", 1, query.Campus).
		Joins("LEFT JOIN peoples AS p ON p.team_id = t.id AND p.open_id = t.captain").
		Where("t.submit = ?", 1)

	if query.ToPointName != "" {
		db = db.Where(
			"EXISTS (SELECT 1 FROM route_edges AS e WHERE e.route_name = t.route_name AND e.prev_point_name = t.prev_point_name AND e.point_name = ?)",
			query.ToPointName,
		)
	}

	if query.PrevPointName != "" {
		db = db.Where("t.prev_point_name = ?", query.PrevPointName)
	}

	if query.Key != "" {
		switch query.SearchType {
		case "team_id":
			db = db.Where("t.id = ?", query.Key)
		case "captain_phone":
			db = db.Where("p.tel = ?", query.Key)
		case "captain_name":
			db = db.Where("p.name = ?", query.Key)
		}
	}

	return db
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
