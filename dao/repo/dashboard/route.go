package repo

import (
	"context"

	"github.com/zjutjh/mygo/ndb"

	"app/dao/query"
)

type DashboardRepo struct {
	query *query.Query
}

func NewDashboardRepo() *DashboardRepo {
	return &DashboardRepo{
		query: query.Use(ndb.Pick()),
	}
}

type RouteNameRow struct {
	Name string `gorm:"column:name"`
}

type RouteStatusCountRow struct {
	RouteName  string `gorm:"column:route_name"`
	WalkStatus string `gorm:"column:walk_status"`
	Count      int64  `gorm:"column:cnt"`
}

type RouteWrongCountRow struct {
	RouteName string `gorm:"column:route_name"`
	Count     int64  `gorm:"column:cnt"`
}

// ListActiveRouteNames 查询启用路线，保证没有报名数据的路线也能返回 0 统计。
func (r *DashboardRepo) ListActiveRouteNames(ctx context.Context) ([]RouteNameRow, error) {
	rows := make([]RouteNameRow, 0)

	err := r.query.Route.WithContext(ctx).
		UnderlyingDB().
		Table("routes").
		Select("name").
		Where("is_active = ?", 1).
		Order("id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ListRouteStatusCounts 查询路线+人员状态聚合，得到总报名与各状态人数。
func (r *DashboardRepo) ListRouteStatusCounts(ctx context.Context) ([]RouteStatusCountRow, error) {
	rows := make([]RouteStatusCountRow, 0)

	err := r.query.People.WithContext(ctx).
		UnderlyingDB().
		Table("peoples AS p").
		Select("t.route_name, p.walk_status, COUNT(1) AS cnt").
		Joins("JOIN teams AS t ON t.id = p.team_id").
		Where("t.submit = ? AND t.route_name IS NOT NULL AND t.route_name <> ''", 1).
		Group("t.route_name, p.walk_status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ListRouteWrongCounts 查询按路线聚合的走错人数。
func (r *DashboardRepo) ListRouteWrongCounts(ctx context.Context) ([]RouteWrongCountRow, error) {
	rows := make([]RouteWrongCountRow, 0)

	err := r.query.People.WithContext(ctx).
		UnderlyingDB().
		Table("peoples AS p").
		Select("t.route_name, COUNT(1) AS cnt").
		Joins("JOIN teams AS t ON t.id = p.team_id").
		Where("t.submit = ? AND t.is_wrong_route = ? AND t.route_name IS NOT NULL AND t.route_name <> ''", 1, 1).
		Group("t.route_name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// CountPeopleOnSegment 统计指定路段上的人数（按 people 计数）。
func (r *DashboardRepo) CountPeopleOnSegment(ctx context.Context, campus string, prevPointName string, toPointName string) (int64, error) {
	filterQuery := TeamFilterQuery{
		Campus:        campus,
		PrevPointName: prevPointName,
		ToPointName:   toPointName,
	}

	var peopleCount int64
	err := r.buildTeamFilterBaseQuery(ctx, filterQuery).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Count(&peopleCount).Error
	if err != nil {
		return 0, err
	}

	return peopleCount, nil
}

// GetCheckpointPeopleCounts 统计点位已到达与未到达人数（按 people 计数）。
func (r *DashboardRepo) GetCheckpointPeopleCounts(ctx context.Context, campus string, pointName string) (passedCount int64, notArrivedCount int64, err error) {
	baseTotal := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams AS t").
		Joins("JOIN routes AS rt ON rt.name = t.route_name AND rt.is_active = ? AND rt.campus = ?", 1, campus).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Where("t.submit = ?", 1).
		Where("EXISTS (SELECT 1 FROM route_edges AS e WHERE e.route_name = t.route_name AND e.point_name = ?)", pointName)

	var totalPeople int64
	err = baseTotal.Count(&totalPeople).Error
	if err != nil {
		return 0, 0, err
	}

	basePassed := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams AS t").
		Joins("JOIN routes AS rt ON rt.name = t.route_name AND rt.is_active = ? AND rt.campus = ?", 1, campus).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Joins("JOIN (SELECT route_name, MAX(seq_order) AS target_seq FROM route_edges WHERE point_name = ? GROUP BY route_name) AS target ON target.route_name = t.route_name", pointName).
		Joins("LEFT JOIN route_edges AS curr ON curr.route_name = t.route_name AND curr.point_name = t.prev_point_name").
		Where("t.submit = ?", 1).
		Where("curr.seq_order >= target.target_seq")

	var passedPeople int64
	err = basePassed.Count(&passedPeople).Error
	if err != nil {
		return 0, 0, err
	}

	notArrivedPeople := totalPeople - passedPeople
	if notArrivedPeople < 0 {
		notArrivedPeople = 0
	}

	return passedPeople, notArrivedPeople, nil
}
