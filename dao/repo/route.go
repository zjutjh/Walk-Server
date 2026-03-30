package repo

import (
	"context"

	"github.com/zjutjh/mygo/ndb"

	"app/dao/query"
)

type RouteRepo struct {
	query *query.Query
}

func NewRouteRepo() *RouteRepo {
	return &RouteRepo{
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

type RoutePointRow struct {
	PointName string `gorm:"column:point_name"`
	SeqOrder  int    `gorm:"column:seq_order"`
}

type WalkStatusCountRow struct {
	WalkStatus string `gorm:"column:walk_status"`
	Count      int64  `gorm:"column:cnt"`
}

type PointPassedCountRow struct {
	PointName string `gorm:"column:point_name"`
	Count     int64  `gorm:"column:cnt"`
}

// ListActiveRouteNames 查询启用路线，保证没有报名数据的路线也能返回 0 统计。
func (r *RouteRepo) ListActiveRouteNames(ctx context.Context) ([]RouteNameRow, error) {
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

// ListActiveRouteNamesByCampus 查询指定校区的启用路线。
func (r *RouteRepo) ListActiveRouteNamesByCampus(ctx context.Context, campus string) ([]RouteNameRow, error) {
	rows := make([]RouteNameRow, 0)

	err := r.query.Route.WithContext(ctx).
		UnderlyingDB().
		Table("routes").
		Select("name").
		Where("is_active = ? AND campus = ?", 1, campus).
		Order("id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ListRouteStatusCounts 查询路线+人员状态聚合，得到总报名与各状态人数。
func (r *RouteRepo) ListRouteStatusCounts(ctx context.Context) ([]RouteStatusCountRow, error) {
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

// ListRouteStatusCountsByCampus 查询指定校区路线+人员状态聚合。
func (r *RouteRepo) ListRouteStatusCountsByCampus(ctx context.Context, campus string) ([]RouteStatusCountRow, error) {
	rows := make([]RouteStatusCountRow, 0)

	err := r.query.People.WithContext(ctx).
		UnderlyingDB().
		Table("peoples AS p").
		Select("t.route_name, p.walk_status, COUNT(1) AS cnt").
		Joins("JOIN teams AS t ON t.id = p.team_id").
		Joins("JOIN routes AS rt ON rt.name = t.route_name AND rt.is_active = ? AND rt.campus = ?", 1, campus).
		Where("t.submit = ?", 1).
		Group("t.route_name, p.walk_status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ListRouteWrongCounts 查询按路线聚合的走错人数。
func (r *RouteRepo) ListRouteWrongCounts(ctx context.Context) ([]RouteWrongCountRow, error) {
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

// ListRouteWrongCountsByCampus 查询指定校区按路线聚合的走错人数。
func (r *RouteRepo) ListRouteWrongCountsByCampus(ctx context.Context, campus string) ([]RouteWrongCountRow, error) {
	rows := make([]RouteWrongCountRow, 0)

	err := r.query.People.WithContext(ctx).
		UnderlyingDB().
		Table("peoples AS p").
		Select("t.route_name, COUNT(1) AS cnt").
		Joins("JOIN teams AS t ON t.id = p.team_id").
		Joins("JOIN routes AS rt ON rt.name = t.route_name AND rt.is_active = ? AND rt.campus = ?", 1, campus).
		Where("t.submit = ? AND t.is_wrong_route = ?", 1, 1).
		Group("t.route_name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ExistsActiveRoute 校验路线是否存在且启用。
func (r *RouteRepo) ExistsActiveRoute(ctx context.Context, routeName string) (bool, error) {
	var total int64
	err := r.query.Route.WithContext(ctx).
		UnderlyingDB().
		Table("routes").
		Where("name = ? AND is_active = ?", routeName, 1).
		Count(&total).Error
	if err != nil {
		return false, err
	}

	return total > 0, nil
}

// ListRoutePoints 查询路线点位顺序。
func (r *RouteRepo) ListRoutePoints(ctx context.Context, routeName string) ([]RoutePointRow, error) {
	rows := make([]RoutePointRow, 0)

	err := r.query.RouteEdge.WithContext(ctx).
		UnderlyingDB().
		Table("route_edges").
		Select("point_name, MIN(seq_order) AS seq_order").
		Where("route_name = ? AND point_name IS NOT NULL AND point_name <> ''", routeName).
		Group("point_name").
		Order("seq_order ASC").
		Order("point_name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ListRoutePointPassedCounts 查询各点位经过人数（按 people 口径）。
func (r *RouteRepo) ListRoutePointPassedCounts(ctx context.Context, routeName string) ([]PointPassedCountRow, error) {
	rows := make([]PointPassedCountRow, 0)

	err := r.query.Checkin.WithContext(ctx).
		UnderlyingDB().
		Raw(
			"SELECT cp.point_name, COUNT(ps.id) AS cnt "+
				"FROM (SELECT DISTINCT team_id, point_name FROM checkins WHERE route_name = ? AND point_name IS NOT NULL AND point_name <> '') AS cp "+
				"JOIN teams AS t ON t.id = cp.team_id AND t.submit = 1 AND t.route_name = ? "+
				"JOIN peoples AS ps ON ps.team_id = t.id "+
				"GROUP BY cp.point_name",
			routeName,
			routeName,
		).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ListSingleRouteStatusCounts 查询单路线的 walk_status 聚合。
func (r *RouteRepo) ListSingleRouteStatusCounts(ctx context.Context, routeName string) ([]WalkStatusCountRow, error) {
	rows := make([]WalkStatusCountRow, 0)

	err := r.query.People.WithContext(ctx).
		UnderlyingDB().
		Table("peoples AS p").
		Select("p.walk_status, COUNT(1) AS cnt").
		Joins("JOIN teams AS t ON t.id = p.team_id").
		Where("t.submit = ? AND t.route_name = ?", 1, routeName).
		Group("p.walk_status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// CountSingleRouteWrongPeople 查询单路线走错人数。
func (r *RouteRepo) CountSingleRouteWrongPeople(ctx context.Context, routeName string) (int64, error) {
	var total int64
	err := r.query.People.WithContext(ctx).
		UnderlyingDB().
		Table("peoples AS p").
		Joins("JOIN teams AS t ON t.id = p.team_id").
		Where("t.submit = ? AND t.route_name = ? AND t.is_wrong_route = ?", 1, routeName, 1).
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

// CountPeopleOnSegment 统计指定路段上的人数（按 people 计数）。
func (r *RouteRepo) CountPeopleOnSegment(ctx context.Context, campus string, prevPointName string, toPointName string) (int64, error) {
	filterQuery := TeamFilterQuery{
		Campus:        campus,
		PrevPointName: prevPointName,
		ToPointName:   toPointName,
	}

	var peopleCount int64
	err := NewTeamRepo().buildTeamFilterBaseQuery(ctx, filterQuery).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Count(&peopleCount).Error
	if err != nil {
		return 0, err
	}

	return peopleCount, nil
}

// GetCheckpointPeopleCounts 统计点位已到达与未到达人数（按 people 计数）。
func (r *RouteRepo) GetCheckpointPeopleCounts(ctx context.Context, campus string, pointName string) (passedCount int64, notArrivedCount int64, err error) {
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
