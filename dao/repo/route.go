package repo

import (
	"context"
	"strings"

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
	SeqOrder  int    `gorm:"column:seq_order"`
	Count     int64  `gorm:"column:cnt"`
}

type StartCheckpointCountRow struct {
	TotalPeople     int64 `gorm:"column:total_people"`
	AbandonedPeople int64 `gorm:"column:abandoned_people"`
	PassedPeople    int64 `gorm:"column:passed_people"`
}

func effectiveWalkStatuses() []string {
	return []string{
		comm.WalkStatusInProgress,
		comm.WalkStatusCompleted,
	}
}

func buildInPlaceholders(size int) string {
	if size <= 0 {
		return ""
	}

	return strings.TrimSuffix(strings.Repeat("?,", size), ",")
}

// IsActiveRouteStartPoint 判断点位是否为指定校区任意启用路线的起点。
func (r *RouteRepo) IsActiveRouteStartPoint(ctx context.Context, campus string, pointName string) (bool, error) {
	var count int64
	err := r.query.RouteEdge.WithContext(ctx).
		UnderlyingDB().
		Raw(
			"SELECT COUNT(DISTINCT rt.name) "+
				"FROM routes AS rt "+
				"JOIN ("+
				"SELECT route_name, MIN(seq_order) AS start_seq "+
				"FROM route_edges "+
				"WHERE point_name IS NOT NULL AND point_name <> '' "+
				"GROUP BY route_name"+
				") AS rs ON rs.route_name = rt.name "+
				"JOIN route_edges AS e ON e.route_name = rt.name AND e.seq_order = rs.start_seq AND e.point_name = ? "+
				"WHERE rt.is_active = 1 AND rt.campus = ?",
			pointName,
			campus,
		).
		Scan(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
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

// ExistsActiveRoute 校验路线是否存在且启用，同时返回所属校区。
func (r *RouteRepo) ExistsActiveRoute(ctx context.Context, routeName string) (campus string, exists bool, err error) {
	var rows []struct {
		Campus string `gorm:"column:campus"`
	}
	err = r.query.Route.WithContext(ctx).
		UnderlyingDB().
		Table("routes").
		Select("campus").
		Where("name = ? AND is_active = ?", routeName, 1).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return "", false, err
	}
	if len(rows) == 0 {
		return "", false, nil
	}
	return rows[0].Campus, true, nil
}

// ListRoutePoints 查询路线点位顺序（按 route_edges 原样返回，不 GROUP BY）。
// 注意：若一条路线的起点和终点是同一个点位名称（如环线），该点位会重复出现。
func (r *RouteRepo) ListRoutePoints(ctx context.Context, routeName string) ([]RoutePointRow, error) {
	rows := make([]RoutePointRow, 0)

	err := r.query.RouteEdge.WithContext(ctx).
		UnderlyingDB().
		Table("route_edges").
		Select("point_name, seq_order").
		Where("route_name = ? AND point_name IS NOT NULL AND point_name <> ''", routeName).
		Order("seq_order ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ListRoutePointPassedCounts 查询各点位累计到达人数（按 people 口径）。
// 统计逻辑：
// 1) 先按队伍在该路线内的**最大** seq_order 计算 reached_seq，避免回扫/补扫导致进度回退。
// 2) 按 reached_seq 聚合有效参与人数（inProgress/completed），再用窗口函数做累计和。
// 注意：seq_order 只在同一 route_name 下比较；不同路线存在相同 seq_order 不影响结果。
// 另外确认了非法打卡也会进入checkins。
// 当前默认业务假设：不考虑"进行中人员半路重组"和"向前异常打卡"场景。
//
// 走错路线口径：队伍按"有效路线"而非报名路线统计。
// 有效路线 = is_wrong_route ? 最新 wrong_route_records.wrong_route_name : route_name。
// 因此走错的队会计入它实际走的路线（如报名 pf-full、错走 pf-half 的队计入 pf-half），
// 既不会在报名线上冻结/虚高，也不会在实走线上隐身。
// 注意：team_reached 的 checkins JOIN 不再按 route_name 过滤——走错打卡的 checkin.route_name
// 仍是报名线，若过滤会在查实走线时被误挡；仅靠 point_name join 到查询线点序表即可正确定位。
func (r *RouteRepo) ListRoutePointPassedCounts(ctx context.Context, routeName string) ([]PointPassedCountRow, error) {
	rows := make([]PointPassedCountRow, 0)
	statuses := effectiveWalkStatuses()
	if len(statuses) == 0 {
		return rows, nil
	}

	statusPlaceholders := buildInPlaceholders(len(statuses))
	// 参数顺序：route_point_seq(?), route_point_seq_dedup(?), team_reached 的有效路线比较(?)
	args := []any{routeName, routeName, routeName}
	for _, status := range statuses {
		args = append(args, status)
	}

	err := r.query.Checkin.WithContext(ctx).
		UnderlyingDB().
		Raw(
			"WITH route_point_seq AS ("+
				"SELECT point_name, seq_order "+
				"FROM route_edges "+
				"WHERE route_name = ? AND point_name IS NOT NULL AND point_name <> '' "+
				"ORDER BY seq_order ASC"+
				"), route_point_seq_dedup AS ("+
				"SELECT point_name, MIN(seq_order) AS seq_order "+
				"FROM route_edges "+
				"WHERE route_name = ? AND point_name IS NOT NULL AND point_name <> '' "+
				"GROUP BY point_name"+
				"), team_reached AS ("+
				"SELECT t.id AS team_id, MAX(rps.seq_order) AS reached_seq "+
				"FROM teams AS t "+
				"JOIN checkins AS c ON c.team_id = t.id AND c.point_name IS NOT NULL AND c.point_name <> '' "+
				"JOIN route_point_seq_dedup AS rps ON rps.point_name = c.point_name "+
				"WHERE t.submit = 1 AND ("+
				"CASE WHEN t.is_wrong_route = 1 THEN COALESCE("+
				"(SELECT w.wrong_route_name FROM wrong_route_records AS w WHERE w.team_id = t.id ORDER BY w.created_at DESC, w.id DESC LIMIT 1), "+
				"t.route_name) ELSE t.route_name END) = ? "+
				"GROUP BY t.id"+
				"), team_people_by_seq AS ("+
				"SELECT tr.reached_seq, COUNT(ps.id) AS people_count "+
				"FROM team_reached AS tr "+
				"JOIN peoples AS ps ON ps.team_id = tr.team_id AND ps.walk_status IN ("+statusPlaceholders+") "+
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
				"SELECT rp.point_name, rp.seq_order, sc.cnt "+
				"FROM route_point_seq AS rp "+
				"JOIN seq_cumulative AS sc ON sc.seq_order = rp.seq_order "+
				"ORDER BY rp.seq_order ASC",
			args...,
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
		Where("t.submit = ? AND t.route_name = ? AND t.is_wrong_route = ?", true, routeName, true).
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

// CountStartCheckpointPeople 统计起点已签到与未到达人数（按 people 计数）。
// 起点未到达口径：报名人数 - 放弃人数 - 该起点签到人数。
// 其中“该起点签到人数”按提交报名队伍在该点存在签到记录、且人员状态为有效参与状态的人数计算。
func (r *RouteRepo) CountStartCheckpointPeople(ctx context.Context, campus string, pointName string) (passedCount int64, notArrivedCount int64, err error) {
	statuses := effectiveWalkStatuses()
	if len(statuses) == 0 {
		return 0, 0, nil
	}

	statusPlaceholders := buildInPlaceholders(len(statuses))
	args := []any{pointName, campus, comm.WalkStatusAbandoned}
	for _, status := range statuses {
		args = append(args, status)
	}
	args = append(args, pointName)

	row := StartCheckpointCountRow{}
	err = r.query.People.WithContext(ctx).
		UnderlyingDB().
		Raw(
			"WITH start_routes AS ("+
				"SELECT rt.name AS route_name "+
				"FROM routes AS rt "+
				"JOIN ("+
				"SELECT route_name, MIN(seq_order) AS start_seq "+
				"FROM route_edges "+
				"WHERE point_name IS NOT NULL AND point_name <> '' "+
				"GROUP BY route_name"+
				") AS rs ON rs.route_name = rt.name "+
				"JOIN route_edges AS e ON e.route_name = rt.name AND e.seq_order = rs.start_seq AND e.point_name = ? "+
				"WHERE rt.is_active = 1 AND rt.campus = ? "+
				"GROUP BY rt.name"+
				") "+
				"SELECT COUNT(DISTINCT ps.id) AS total_people, "+
				"COUNT(DISTINCT CASE WHEN ps.walk_status = ? THEN ps.id END) AS abandoned_people, "+
				"COUNT(DISTINCT CASE WHEN ps.walk_status IN ("+statusPlaceholders+") AND EXISTS ("+
				"SELECT 1 FROM checkins AS c WHERE c.team_id = t.id AND c.point_name = ?"+
				") THEN ps.id END) AS passed_people "+
				"FROM teams AS t "+
				"JOIN start_routes AS sr ON sr.route_name = t.route_name "+
				"JOIN peoples AS ps ON ps.team_id = t.id "+
				"WHERE t.submit = 1",
			args...,
		).
		Scan(&row).Error
	if err != nil {
		return 0, 0, err
	}

	notArrivedPeople := row.TotalPeople - row.AbandonedPeople - row.PassedPeople
	if notArrivedPeople < 0 {
		notArrivedPeople = 0
	}

	return row.PassedPeople, notArrivedPeople, nil
}

// CountPeopleOnSegment 统计指定路段上的人数（按 people 计数）。
// 口径说明：统计"队伍当前 latest_point_name + 有效路线"能匹配到该边且有效成员（进行中）的人数。
// 有效路线 = is_wrong_route ? 最新 wrong_route_records.wrong_route_name : route_name，与 buildTeamFilterBaseQuery 一致。
func (r *RouteRepo) CountPeopleOnSegment(ctx context.Context, campus string, prevPointName string, toPointName string) (int64, error) {
	filterQuery := TeamFilterQuery{
		Campus:        campus,
		PrevPointName: prevPointName,
		ToPointName:   toPointName,
	}

	var peopleCount int64
	err := NewTeamRepo().buildTeamFilterBaseQuery(ctx, filterQuery).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Where("ps.walk_status = ?", comm.WalkStatusInProgress).
		Count(&peopleCount).Error
	if err != nil {
		return 0, err
	}

	return peopleCount, nil
}

// GetCheckpointPeopleCounts 统计点位已到达与未到达人数（按 people 计数）。
// 口径说明：已到达判断基于 teams.latest_point_name 在所属路线上的 seq_order 与目标点序比较。
// 若队伍当前点无法映射到所属路线（如错路期间打到他路线独有点），该队会被视为"未到达"。
func (r *RouteRepo) GetCheckpointPeopleCounts(ctx context.Context, campus string, pointName string) (passedCount int64, notArrivedCount int64, err error) {
	statuses := effectiveWalkStatuses()
	if len(statuses) == 0 {
		return 0, 0, nil
	}

	isStartPoint, err := r.IsActiveRouteStartPoint(ctx, campus, pointName)
	if err != nil {
		return 0, 0, err
	}
	if isStartPoint {
		return r.CountStartCheckpointPeople(ctx, campus, pointName)
	}

	baseTotal := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams AS t").
		Joins("JOIN routes AS rt ON rt.name = t.route_name AND rt.is_active = ? AND rt.campus = ?", 1, campus).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Where("t.submit = ?", 1).
		Where("ps.walk_status IN ?", statuses).
		Where("EXISTS (SELECT 1 FROM route_edges AS e WHERE e.route_name = (CASE WHEN t.is_wrong_route = 1 THEN COALESCE((SELECT w.wrong_route_name FROM wrong_route_records AS w WHERE w.team_id = t.id ORDER BY w.created_at DESC, w.id DESC LIMIT 1), t.route_name) ELSE t.route_name END) AND e.point_name = ?)", pointName)

	var totalPeople int64
	err = baseTotal.Distinct("ps.id").Count(&totalPeople).Error
	if err != nil {
		return 0, 0, err
	}

	basePassed := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams AS t").
		Joins("JOIN routes AS rt ON rt.name = t.route_name AND rt.is_active = ? AND rt.campus = ?", 1, campus).
		Joins("JOIN peoples AS ps ON ps.team_id = t.id").
		Joins("JOIN (SELECT route_name, MIN(seq_order) AS target_seq FROM route_edges WHERE point_name = ? GROUP BY route_name) AS target ON target.route_name = (CASE WHEN t.is_wrong_route = 1 THEN COALESCE((SELECT w.wrong_route_name FROM wrong_route_records AS w WHERE w.team_id = t.id ORDER BY w.created_at DESC, w.id DESC LIMIT 1), t.route_name) ELSE t.route_name END)", pointName).
		Joins("LEFT JOIN (SELECT route_name, point_name, MIN(seq_order) AS seq_order FROM route_edges GROUP BY route_name, point_name) AS curr ON curr.route_name = (CASE WHEN t.is_wrong_route = 1 THEN COALESCE((SELECT w.wrong_route_name FROM wrong_route_records AS w WHERE w.team_id = t.id ORDER BY w.created_at DESC, w.id DESC LIMIT 1), t.route_name) ELSE t.route_name END) AND curr.point_name = t.latest_point_name").
		Where("t.submit = ?", 1).
		Where("ps.walk_status IN ?", statuses).
		Where("curr.seq_order >= target.target_seq")

	var passedPeople int64
	err = basePassed.Distinct("ps.id").Count(&passedPeople).Error
	if err != nil {
		return 0, 0, err
	}

	notArrivedPeople := totalPeople - passedPeople
	if notArrivedPeople < 0 {
		notArrivedPeople = 0
	}

	return passedPeople, notArrivedPeople, nil
}
