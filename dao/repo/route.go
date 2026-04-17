package repo

import (
	"context"

	"app/comm"

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

func effectiveWalkStatuses() []string {
	return []string{
		comm.WalkStatusInProgress,
		comm.WalkStatusCompleted,
		comm.WalkStatusViolated,
	}
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

// ListRoutePointPassedCounts 查询各点位累计到达人数（按 people 口径）。
// 统计逻辑：
// 1) 先按队伍在该路线内的**最大** seq_order 计算 reached_seq，避免回扫/补扫导致进度回退。
// 2) 按 reached_seq 聚合有效参与人数（inProgress/completed/violated），再用窗口函数做累计和。
// 注意：seq_order 只在同一 route_name 下比较；不同路线存在相同 seq_order 不影响结果。
// 另外确认了非法打卡也会进入checkins。
// 当前默认业务假设：不考虑“进行中人员半路重组”和“向前异常打卡”场景。
func (r *RouteRepo) ListRoutePointPassedCounts(ctx context.Context, routeName string) ([]PointPassedCountRow, error) {
	rows := make([]PointPassedCountRow, 0)
	statuses := effectiveWalkStatuses()

	err := r.query.Checkin.WithContext(ctx).
		UnderlyingDB().
		Raw(
			"WITH route_point_seq AS ("+
				"SELECT point_name, MIN(seq_order) AS seq_order "+
				"FROM route_edges "+
				"WHERE route_name = ? AND point_name IS NOT NULL AND point_name <> '' "+
				"GROUP BY point_name"+
				"), team_reached AS ("+
				"SELECT t.id AS team_id, MAX(rps.seq_order) AS reached_seq "+
				"FROM teams AS t "+
				"JOIN checkins AS c ON c.team_id = t.id AND c.route_name = ? AND c.point_name IS NOT NULL AND c.point_name <> '' "+
				"JOIN route_point_seq AS rps ON rps.point_name = c.point_name "+
				"WHERE t.submit = 1 AND t.route_name = ? "+
				"GROUP BY t.id"+
				"), team_people_by_seq AS ("+
				"SELECT tr.reached_seq, COUNT(ps.id) AS people_count "+
				"FROM team_reached AS tr "+
				"JOIN peoples AS ps ON ps.team_id = tr.team_id AND ps.walk_status IN (?, ?, ?) "+
				"GROUP BY tr.reached_seq"+
				"), seq_levels AS ("+
				"SELECT DISTINCT seq_order FROM route_point_seq"+
				"), seq_cumulative AS ("+
				"SELECT sl.seq_order, COALESCE(SUM(COALESCE(tps.people_count, 0)) OVER ("+
				"ORDER BY sl.seq_order DESC ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW"+
				"), 0) AS cnt "+
				"FROM seq_levels AS sl "+
				"LEFT JOIN team_people_by_seq AS tps ON tps.reached_seq = sl.seq_order"+
				") "+
				"SELECT rp.point_name, sc.cnt "+
				"FROM route_point_seq AS rp "+
				"JOIN seq_cumulative AS sc ON sc.seq_order = rp.seq_order "+
				"ORDER BY rp.seq_order ASC, rp.point_name ASC",
			routeName,
			routeName,
			routeName,
			statuses[0],
			statuses[1],
			statuses[2],
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
// 口径说明：仅统计“队伍当前 prev_point_name + 队伍所属 route_name”能匹配到该边的人数。
// 若队伍发生错路且当前点不在所属路线边上，该队会暂时不计入任何标准路段，直到回到可匹配边。
// 换句话说，就是进度在人数统计层面会短暂不更新，即使已经打卡，直到回到正确路线（报名的路线）的边上。
func (r *RouteRepo) CountPeopleOnSegment(ctx context.Context, campus string, prevPointName string, toPointName string) (int64, error) {
	filterQuery := TeamFilterQuery{
		Campus:        campus,
		PrevPointName: prevPointName,
		ToPointName:   toPointName,
	}
	statuses := effectiveWalkStatuses()

	var peopleCount int64
	err := NewTeamRepo().buildTeamFilterBaseQuery(ctx, filterQuery).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Where("ps.walk_status IN ?", statuses).
		Count(&peopleCount).Error
	if err != nil {
		return 0, err
	}

	return peopleCount, nil
}

// GetCheckpointPeopleCounts 统计点位已到达与未到达人数（按 people 计数）。
// 口径说明：已到达判断基于 teams.prev_point_name 在所属路线上的 seq_order 与目标点序比较。
// 若队伍当前点无法映射到所属路线（如错路期间打到他路线独有点），该队会被视为“未到达”。
func (r *RouteRepo) GetCheckpointPeopleCounts(ctx context.Context, campus string, pointName string) (passedCount int64, notArrivedCount int64, err error) {
	statuses := effectiveWalkStatuses()

	baseTotal := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams AS t").
		Joins("JOIN routes AS rt ON rt.name = t.route_name AND rt.is_active = ? AND rt.campus = ?", 1, campus).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Where("t.submit = ?", 1).
		Where("ps.walk_status IN ?", statuses).
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
		Where("ps.walk_status IN ?", statuses).
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
