package repo

import (
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
	TeamID        int64        `gorm:"column:team_id"`
	CaptainName   string       `gorm:"column:captain_name"`
	CaptainPhone  string       `gorm:"column:captain_phone"`
	PrevPointName string       `gorm:"column:prev_point_name"`
	PrevPointTime sql.NullTime `gorm:"column:prev_point_time"`
	RouteName     string       `gorm:"column:route_name"`
	IsLost        int8         `gorm:"column:is_lost"`
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
	_ = teamCache.DelTeamByID(ctx, id)
	return nil
}

func (r *TeamRepo) IncrementNumIfAvailable(ctx context.Context, id int64, maxTeamSize int) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND submit = ? AND num < ?", id, 0, maxTeamSize).
		UpdateColumn("num", gorm.Expr("num + ?", 1))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		_ = teamCache.DelTeamByID(ctx, id)
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DecrementNumIfPositive(ctx context.Context, id int64) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Team{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND submit = ? AND num > 0", id, 0).
		UpdateColumn("num", gorm.Expr("num - ?", 1))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		_ = teamCache.DelTeamByID(ctx, id)
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
	_ = teamCache.DelTeamByID(ctx, id)
	return nil
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

func (r *TeamRepo) UpdateTeamWrongRoute(ctx context.Context, teamID int64, isWrongRoute int8) error {
	t := r.query.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Update(t.IsWrongRoute, isWrongRoute)
	if err == nil {
		_ = teamCache.DelTeamByID(ctx, teamID)
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
			t.IsLost.Eq(1),
		).
		Update(t.IsLost, 0)
	if err == nil {
		_ = teamCache.DelTeamByID(ctx, teamID)
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

func (r *TeamRepo) FindRouteEdge(ctx context.Context, routeName, pointName string) (*model.RouteEdge, error) {
	if routeEdge, hit, err := routeCache.GetRouteEdge(ctx, routeName, pointName); err == nil && hit {
		return routeEdge, nil
	}

	re := r.query.RouteEdge
	record, err := re.WithContext(ctx).
		Where(
			re.RouteName.Eq(routeName),
			re.PointName.Eq(pointName),
		).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = routeCache.SetRouteEdge(ctx, record)
	return record, nil
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

func (r *TeamRepo) UpdatePrevPointName(ctx context.Context, teamID int64, pointName string) error {
	t := r.query.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Updates(map[string]any{
			"prev_point_name": pointName,
			"time":            time.Now(),
		})
	if err == nil {
		_ = teamCache.DelTeamByID(ctx, teamID)
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

func (r *TeamRepo) buildTeamFilterBaseQuery(ctx context.Context, query TeamFilterQuery) *gorm.DB {
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

func (r *TeamRepo) UpdateTeamLostStatus(ctx context.Context, teamID int64, isLost bool, statusUpdatedAt time.Time) (bool, error) {
	isLostVal := int8(0)
	if isLost {
		isLostVal = 1
	}

	m := map[string]interface{}{
		"is_lost": isLostVal,
	}

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
	if tx.RowsAffected > 0 {
		_ = teamCache.DelTeamByID(ctx, teamID)
	}

	return tx.RowsAffected > 0, nil
}
