package repo

import (
	"app/comm"
	routeCache "app/dao/cache/route"
	teamCache "app/dao/cache/team"
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"app/dao/model"
	"app/dao/query"
)

type TeamRepo struct {
	db    *gorm.DB
	query *query.Query
}

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
	TeamID          int64        `gorm:"column:team_id"`
	CaptainName     string       `gorm:"column:captain_name"`
	CaptainPhone    string       `gorm:"column:captain_phone"`
	LatestPointName string       `gorm:"column:latest_point_name"`
	LatestPointTime sql.NullTime `gorm:"column:latest_point_time"`
	RouteName       string       `gorm:"column:route_name"`
	IsLost          bool         `gorm:"column:is_lost"`
}

type TeamCheckinRow struct {
	ID        int64     `gorm:"column:id"`
	AdminID   int64     `gorm:"column:admin_id"`
	TeamID    int64     `gorm:"column:team_id"`
	PointName string    `gorm:"column:point_name"`
	RouteName string    `gorm:"column:route_name"`
	Time      time.Time `gorm:"column:time"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func invalidateTeamCaches(ctx context.Context, teamID int64) {
	_ = teamCache.DelTeamByID(ctx, teamID)
	_ = teamCache.DeleteTeamInfo(ctx, teamID)
}

func NewTeamRepo() *TeamRepo {
	db := ndb.Pick()
	return &TeamRepo{
		db:    db,
		query: query.Use(db),
	}

}

func NewTeamRepoWithTx(tx *query.Query) *TeamRepo {
	return &TeamRepo{
		db:    tx.Team.WithContext(context.Background()).UnderlyingDB(),
		query: tx,
	}
}

func (r *TeamRepo) Create(ctx context.Context, team *model.Team) error {
	if err := r.query.Team.WithContext(ctx).Create(team); err != nil {
		return err
	}
	_ = teamCache.SetTeamByID(ctx, team)
	if team.Code != "" {
		_ = teamCache.SetTeamIDByCode(ctx, team.Code, team.ID)
	}
	return nil
}

// FindTeamByID 根据ID查询队伍
func (r *TeamRepo) FindTeamByID(ctx context.Context, id int64) (*model.Team, error) {
	if team, hit, err := teamCache.GetTeamByID(ctx, id); err == nil && hit {
		return team, nil
	}

	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.ID.Eq(id)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = teamCache.SetTeamByID(ctx, record)
	if record.Code != "" {
		_ = teamCache.SetTeamIDByCode(ctx, record.Code, record.ID)
	}
	return record, nil
}

func (r *TeamRepo) GetTeamByID(ctx context.Context, teamID int64) (*model.Team, error) {
	return r.query.Team.WithContext(ctx).
		Where(r.query.Team.ID.Eq(teamID)).
		First()
}

func (r *TeamRepo) FindTeamByName(ctx context.Context, name string) (*model.Team, error) {
	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.Name.Eq(name)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if record.Code != "" {
		_ = teamCache.SetTeamIDByCode(ctx, record.Code, record.ID)
	}
	return record, nil
}

func (r *TeamRepo) FindByNameExceptID(ctx context.Context, name string, id int64) (*model.Team, error) {
	t := r.query.Team
	record, err := t.WithContext(ctx).
		Where(
			t.Name.Eq(name),
			t.ID.Neq(id),
		).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if record.Code != "" {
		_ = teamCache.SetTeamIDByCode(ctx, record.Code, record.ID)
	}
	return record, nil
}

func (r *TeamRepo) FindByCode(ctx context.Context, code string) (*model.Team, error) {
	if teamID, hit, err := teamCache.GetTeamIDByCode(ctx, code); err == nil && hit {
		return r.FindTeamByID(ctx, teamID)
	}

	t := r.query.Team
	record, err := t.WithContext(ctx).Where(t.Code.Eq(code)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = teamCache.SetTeamIDByCode(ctx, record.Code, record.ID)
	return record, nil
}

func (r *TeamRepo) UpdateByID(ctx context.Context, id int64, updates map[string]any) error {
	_, err := r.query.Team.WithContext(ctx).
		Where(r.query.Team.ID.Eq(id)).
		Updates(updates)
	if err != nil {
		return err
	}
	invalidateTeamCaches(ctx, id)
	return nil
}

func (r *TeamRepo) UpdateStatusByIDs(ctx context.Context, ids []int64, status string) error {
	if len(ids) == 0 {
		return nil
	}
	t := r.query.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.In(ids...)).
		Update(t.Status, status)
	if err != nil {
		return err
	}
	for _, id := range ids {
		invalidateTeamCaches(ctx, id)
	}
	return nil
}

func (r *TeamRepo) IncrementNumIfAvailable(ctx context.Context, id int64, maxTeamSize int) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND num < ?", id, maxTeamSize).
		UpdateColumn("num", gorm.Expr("num + ?", 1))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		invalidateTeamCaches(ctx, id)
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DecrementNumIfPositive(ctx context.Context, id int64) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND num > 0", id).
		UpdateColumn("num", gorm.Expr("num - ?", 1))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		invalidateTeamCaches(ctx, id)
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.query.Team.WithContext(ctx).
		Where(r.query.Team.ID.Eq(id)).
		Delete()
	if err != nil {
		return err
	}
	invalidateTeamCaches(ctx, id)
	return nil
}

func (r *TeamRepo) CreateWithCaptain(ctx context.Context, team *model.Team, captain *model.People) error {
	return query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		teamRepo := NewTeamRepoWithTx(tx)
		peopleRepo := NewPeopleRepoWithTx(tx)
		if err := teamRepo.Create(ctx, team); err != nil {
			return err
		}
		return peopleRepo.UpdateByOpenID(ctx, captain.OpenID, map[string]any{
			"created_op": captain.CreatedOp - 1,
			"role":       comm.RoleCaptain,
			"team_id":    team.ID,
		})
	})
}

func (r *TeamRepo) JoinTeam(ctx context.Context, teamID int64, person *model.People, consumeJoinOp bool, maxTeamSize int) (bool, error) {
	updates := map[string]any{
		"role":    comm.RoleMember,
		"team_id": teamID,
	}
	if consumeJoinOp {
		updates["join_op"] = person.JoinOp - 1
	}

	joined := false
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		teamRepo := NewTeamRepoWithTx(tx)
		ok, err := teamRepo.IncrementNumIfAvailable(ctx, teamID, maxTeamSize)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if err := NewPeopleRepoWithTx(tx).UpdateByOpenID(ctx, person.OpenID, updates); err != nil {
			return err
		}
		joined = true
		return nil
	})
	return joined, err
}

func (r *TeamRepo) RemoveMember(ctx context.Context, teamID int64, person *model.People) (bool, error) {
	removed := false
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		teamRepo := NewTeamRepoWithTx(tx)
		ok, err := teamRepo.DecrementNumIfPositive(ctx, teamID)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if err := NewPeopleRepoWithTx(tx).UpdateByOpenID(ctx, person.OpenID, map[string]any{
			"role":    comm.RoleUnbind,
			"team_id": int64(-1),
		}); err != nil {
			return err
		}
		removed = true
		return nil
	})
	return removed, err
}

func (r *TeamRepo) ChangeCaptain(ctx context.Context, teamID int64, oldCaptainOpenID, newCaptainOpenID string) error {
	return query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		teamRepo := NewTeamRepoWithTx(tx)
		peopleRepo := NewPeopleRepoWithTx(tx)
		if err := teamRepo.UpdateByID(ctx, teamID, map[string]any{"captain": newCaptainOpenID}); err != nil {
			return err
		}
		if err := peopleRepo.UpdateByOpenID(ctx, oldCaptainOpenID, map[string]any{"role": comm.RoleMember}); err != nil {
			return err
		}
		return peopleRepo.UpdateByOpenID(ctx, newCaptainOpenID, map[string]any{"role": comm.RoleCaptain})
	})
}

func (r *TeamRepo) DisbandTeam(ctx context.Context, teamID int64) error {
	return query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		if err := NewPeopleRepoWithTx(tx).UpdateByTeamID(ctx, teamID, map[string]any{
			"role":    comm.RoleUnbind,
			"team_id": int64(-1),
		}); err != nil {
			return err
		}
		return NewTeamRepoWithTx(tx).DeleteByID(ctx, teamID)
	})
}

func (r *TeamRepo) ListRandomMatchTeams(ctx context.Context, routeName string, maxTeamSize int) ([]model.Team, error) {
	teams := make([]model.Team, 0)
	tiers := []struct {
		where string
		limit int
	}{
		{where: "num <= 3", limit: 3},
		{where: "num = 4", limit: 4},
		{where: "num = 5", limit: 5},
	}
	for _, tier := range tiers {
		if len(teams) >= 5 {
			break
		}
		var rows []model.Team
		limit := tier.limit - len(teams)
		if limit <= 0 {
			continue
		}
		err := r.db.WithContext(ctx).
			Model(&model.Team{}).
			Where("route_name = ? AND allow_match = ? AND num < ?", routeName, true, maxTeamSize).
			Where(tier.where).
			Order("RAND()").
			Limit(limit).
			Find(&rows).Error
		if err != nil {
			return nil, err
		}
		teams = append(teams, rows...)
	}
	return teams, nil
}

func (r *TeamRepo) CreateMessage(ctx context.Context, senderID *int64, receiverID int64, message string) error {
	return r.db.WithContext(ctx).
		Exec("INSERT INTO messages (sender_id, receiver_id, message) VALUES (?, ?, ?)", senderID, receiverID, message).
		Error
}

func (r *TeamRepo) CreateCheckin(ctx context.Context, adminID, teamID int64, pointName, routeName string) error {
	checkin := &model.Checkin{
		AdminID:   adminID,
		TeamID:    teamID,
		PointName: pointName,
		RouteName: routeName,
		Time:      time.Now(),
	}
	return r.query.Checkin.WithContext(ctx).Create(checkin)
}

func (r *TeamRepo) UpdateTeamWrongRoute(ctx context.Context, teamID int64, isWrongRoute bool) error {
	t := r.query.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.IsWrongRoute, isWrongRoute)
	if err == nil {
		invalidateTeamCaches(ctx, teamID)
	}
	return err
}

func (r *TeamRepo) CreateWrongRouteRecord(ctx context.Context, teamID int64, routeName, wrongRouteName string, adminID int64) error {
	record := &model.WrongRouteRecord{
		TeamID:         teamID,
		RouteName:      routeName,
		WrongRouteName: wrongRouteName,
		AdminID:        adminID,
	}
	return r.query.WrongRouteRecord.WithContext(ctx).Create(record)
}

func (r *TeamRepo) ClearLostStatus(ctx context.Context, teamID int64) error {
	t := r.query.Team
	_, err := t.WithContext(ctx).
		Where(
			t.ID.Eq(teamID),
			t.IsLost.Is(true),
		).
		Update(t.IsLost, false)
	if err == nil {
		invalidateTeamCaches(ctx, teamID)
	}
	return err
}

func (r *TeamRepo) FindRouteByName(ctx context.Context, routeName string) (*model.Route, error) {
	if route, hit, err := routeCache.GetRoute(ctx, routeName); err == nil && hit {
		return route, nil
	}

	rt := r.query.Route
	record, err := rt.WithContext(ctx).Where(rt.Name.Eq(routeName)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = routeCache.SetRoute(ctx, record)
	return record, nil
}

func (r *TeamRepo) FindRouteTransitionEdge(ctx context.Context, routeName, prevPointName, pointName string) (*model.RouteEdge, error) {
	if routeEdge, hit, err := routeCache.GetRouteEdge(ctx, routeName, prevPointName, pointName); err == nil && hit {
		return routeEdge, nil
	}

	re := r.query.RouteEdge
	query := re.WithContext(ctx).Where(
		re.RouteName.Eq(routeName),
		re.PointName.Eq(pointName),
	)
	if prevPointName == "" {
		query = query.Where(re.PrevPointName.IsNull())
	} else {
		query = query.Where(re.PrevPointName.Eq(prevPointName))
	}

	record, err := query.First()
	if err == nil {
		_ = routeCache.SetRouteEdge(ctx, record)
		return record, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return nil, nil
}

func (r *TeamRepo) FindPointRoutes(ctx context.Context, pointName string) ([]string, error) {
	if routeNames, hit, err := routeCache.GetPointRoutes(ctx, pointName); err == nil && hit {
		return routeNames, nil
	}

	re := r.query.RouteEdge
	var routeNames []string
	err := re.WithContext(ctx).
		Where(re.PointName.Eq(pointName)).
		Pluck(re.RouteName, &routeNames)
	if err != nil {
		return nil, err
	}
	_ = routeCache.SetPointRoutes(ctx, pointName, routeNames)
	return routeNames, nil
}

func (r *TeamRepo) IsPointOnRoute(ctx context.Context, routeName, pointName string) (bool, error) {
	var total int64
	err := r.query.RouteEdge.WithContext(ctx).
		UnderlyingDB().
		Table("route_edges").
		Where("route_name = ? AND point_name = ?", routeName, pointName).
		Count(&total).Error
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *TeamRepo) IsRouteTransitionValid(ctx context.Context, routeName, prevPointName, pointName string) (bool, error) {
	var total int64
	err := r.query.RouteEdge.WithContext(ctx).
		UnderlyingDB().
		Table("route_edges").
		Where("route_name = ? AND prev_point_name = ? AND point_name = ?", routeName, prevPointName, pointName).
		Count(&total).Error
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *TeamRepo) IsDirectionBackward(ctx context.Context, routeName, prevPointName, pointName string) (bool, error) {
	if prevPointName == "" {
		return false, nil
	}

	isDirectNext, err := r.IsRouteTransitionValid(ctx, routeName, prevPointName, pointName)
	if err != nil {
		return false, err
	}
	if isDirectNext {
		return false, nil
	}

	var nextSeq struct {
		SeqOrder int `gorm:"column:seq_order"`
	}
	err = r.query.RouteEdge.WithContext(ctx).
		UnderlyingDB().
		Table("route_edges").
		Select("MIN(seq_order) AS seq_order").
		Where("route_name = ? AND prev_point_name = ?", routeName, prevPointName).
		Scan(&nextSeq).Error
	if err != nil {
		return false, err
	}
	if nextSeq.SeqOrder == 0 {
		return false, nil
	}

	var currentSeq struct {
		SeqOrder int `gorm:"column:seq_order"`
	}
	err = r.query.RouteEdge.WithContext(ctx).
		UnderlyingDB().
		Table("route_edges").
		Select("MAX(seq_order) AS seq_order").
		Where("route_name = ? AND point_name = ?", routeName, pointName).
		Scan(&currentSeq).Error
	if err != nil {
		return false, err
	}
	if currentSeq.SeqOrder == 0 {
		return false, nil
	}
	return currentSeq.SeqOrder < nextSeq.SeqOrder, nil
}

func (r *TeamRepo) ListLatestCheckins(ctx context.Context, teamID int64, limit int) ([]TeamCheckinRow, error) {
	rows := make([]TeamCheckinRow, 0, limit)
	if limit <= 0 {
		return rows, nil
	}
	err := r.query.Checkin.WithContext(ctx).
		UnderlyingDB().
		Table("checkins").
		Select("id, admin_id, team_id, point_name, route_name, time, created_at").
		Where("team_id = ?", teamID).
		Order("time DESC").
		Order("id DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *TeamRepo) GetLatestWrongRouteName(ctx context.Context, teamID int64) (string, bool, error) {
	var row struct {
		WrongRouteName string `gorm:"column:wrong_route_name"`
	}
	err := r.query.WrongRouteRecord.WithContext(ctx).
		UnderlyingDB().
		Table("wrong_route_records").
		Select("wrong_route_name").
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Order("id DESC").
		Limit(1).
		Scan(&row).Error
	if err != nil {
		return "", false, err
	}
	if row.WrongRouteName == "" {
		return "", false, nil
	}
	return row.WrongRouteName, true, nil
}

func (r *TeamRepo) UpdateLatestPointName(ctx context.Context, teamID int64, pointName string) error {
	t := r.query.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Updates(map[string]any{
			"latest_point_name": pointName,
			"time":              time.Now(),
		})
	if err == nil {
		invalidateTeamCaches(ctx, teamID)
	}
	return err
}

func (r *TeamRepo) ListTeamMembers(ctx context.Context, teamID int64) ([]TeamMemberRow, error) {
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

func roleSortRank(role string) int {
	if strings.EqualFold(role, "captain") {
		return 0
	}

	return 1
}

func (r *TeamRepo) CountTeamsByFilter(ctx context.Context, query TeamFilterQuery) (int64, error) {
	var total int64
	err := r.buildTeamFilterBaseQuery(ctx, query).
		Distinct("t.id").
		Count(&total).Error
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *TeamRepo) ListTeamsByFilter(ctx context.Context, query TeamFilterQuery) ([]TeamFilterRow, error) {
	rows := make([]TeamFilterRow, 0)

	err := r.buildTeamFilterBaseQuery(ctx, query).
		Select("t.id AS team_id, COALESCE(p.name, '') AS captain_name, COALESCE(p.tel, '') AS captain_phone, t.latest_point_name, t.time AS latest_point_time, t.route_name, t.is_lost").
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

func (r *TeamRepo) buildTeamFilterBaseQuery(ctx context.Context, query TeamFilterQuery) *gorm.DB {
	db := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams AS t").
		Joins("JOIN routes AS r ON r.name = t.route_name AND r.is_active = ? AND r.campus = ?", 1, query.Campus).
		Joins("LEFT JOIN peoples AS p ON p.team_id = t.id AND p.open_id = t.captain").
		Where("t.submit = ?", 1)

	effRoute := "(CASE WHEN t.is_wrong_route = 1 THEN COALESCE((SELECT w.wrong_route_name FROM wrong_route_records AS w WHERE w.team_id = t.id ORDER BY w.created_at DESC, w.id DESC LIMIT 1), t.route_name) ELSE t.route_name END)"

	if query.ToPointName != "" && query.PrevPointName != "" {
		db = db.Where(
			"EXISTS (SELECT 1 FROM route_edges AS e WHERE e.route_name = "+effRoute+" AND e.prev_point_name = ? AND e.point_name = ?)",
			query.PrevPointName,
			query.ToPointName,
		)
	} else if query.ToPointName != "" {
		db = db.Where(
			"EXISTS (SELECT 1 FROM route_edges AS e WHERE e.route_name = "+effRoute+" AND e.prev_point_name = t.latest_point_name AND e.point_name = ?)",
			query.ToPointName,
		)
		// 路段筛选（指定了终点）时，剔除"待出发队伍"——即全队没有任何成员处于有效行进状态
		// （in_progress/violated）。口径与 route/segment 的人数统计一致，避免
		// 尚未真正出发的队伍出现在路段队伍列表里。路段为空（仅关键词搜索）时不施加此过滤。
		db = db.Where(
			"EXISTS (SELECT 1 FROM peoples AS ps WHERE ps.team_id = t.id AND ps.walk_status IN ?)",
			[]string{"in_progress", "violated"},
		)
	}

	if query.PrevPointName != "" {
		db = db.Where("t.latest_point_name = ?", query.PrevPointName)
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

func (r *TeamRepo) UpdateTeamLostStatus(ctx context.Context, teamID int64, isLost bool) (bool, error) {
	m := map[string]interface{}{
		"is_lost": isLost,
	}

	tx := r.query.Team.WithContext(ctx).
		UnderlyingDB().
		Table("teams").
		Where("id = ?", teamID).
		Updates(m)
	if tx.Error != nil {
		return false, tx.Error
	}
	if tx.RowsAffected > 0 {
		invalidateTeamCaches(ctx, teamID)
	}

	return tx.RowsAffected > 0, nil
}
