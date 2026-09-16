package repo

import (
	"app/comm"
	routeCache "app/dao/cache/route"
	teamCache "app/dao/cache/team"
	"context"
	"database/sql"
	"errors"
	"math/rand"
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
	query *query.Query
}

type TeamMemberRow struct {
	ID    int64
	Name  string
	Phone string
	Role  string
}

type teamMatchCountRow struct {
	Num   uint8 `gorm:"column:num"`
	Count int64 `gorm:"column:count"`
}

func randomOffset(count int64, take int) int {
	if count <= int64(take) {
		return 0
	}
	return rand.Intn(int(count) - take + 1)
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

func NewTeamRepo() *TeamRepo {
	return &TeamRepo{query: query.Use(ndb.Pick())}
}

func NewTeamRepoWithTx(tx *query.Query) *TeamRepo {
	return &TeamRepo{query: tx}
}

func (r *TeamRepo) Create(ctx context.Context, team *model.Team) error {
	teamQuery := r.query.Team.WithContext(ctx)
	if team.Code == "" {
		teamQuery = teamQuery.Omit(r.query.Team.Code)
	}
	return teamQuery.Create(team)
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
		team, err := r.FindTeamByID(ctx, teamID)
		if err != nil {
			return nil, err
		}
		if team != nil && team.Code == code {
			return team, nil
		}
		_ = teamCache.DelTeamIDByCode(ctx, code)
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

func (r *TeamRepo) UpdateCodeByID(ctx context.Context, id int64, code string) error {
	_, err := r.query.Team.WithContext(ctx).
		Where(r.query.Team.ID.Eq(id)).
		Update(r.query.Team.Code, code)
	return err
}

func (r *TeamRepo) UpdateByID(ctx context.Context, id int64, updates map[string]any) error {
	_, err := r.query.Team.WithContext(ctx).
		Where(r.query.Team.ID.Eq(id)).
		Updates(updates)
	return err
}

func (r *TeamRepo) UpdateWithNotices(ctx context.Context, id int64, updates map[string]any, userIDs []int64, noticeTypes []comm.NoticeType) error {
	return r.query.Transaction(func(tx *query.Query) error {
		if err := NewTeamRepoWithTx(tx).UpdateByID(ctx, id, updates); err != nil {
			return err
		}
		notices := make([]*model.Notice, 0, len(userIDs)*len(noticeTypes))
		for _, userID := range userIDs {
			for _, noticeType := range noticeTypes {
				teamID := id
				notices = append(notices, &model.Notice{UserID: userID, Type: noticeType, TeamID: &teamID})
			}
		}
		return NewNoticeRepoWithTx(tx).UpsertUnread(ctx, notices)
	})
}

func (r *TeamRepo) UpdateStatusByIDs(ctx context.Context, ids []int64, status string) error {
	if len(ids) == 0 {
		return nil
	}
	t := r.query.Team
	_, err := t.WithContext(ctx).
		Where(t.ID.In(ids...)).
		Update(t.Status, status)
	return err
}

func (r *TeamRepo) IncrementNumIfAvailable(ctx context.Context, id int64, maxTeamSize int) (bool, error) {
	t := r.query.Team
	result, err := t.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(t.ID.Eq(id), t.Num.Lt(uint8(maxTeamSize))).
		UpdateColumn(t.Num, gorm.Expr("num + ?", 1))
	if err != nil {
		return false, err
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DecrementNumIfPositive(ctx context.Context, id int64) (bool, error) {
	t := r.query.Team
	result, err := t.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(t.ID.Eq(id), t.Num.Gt(0)).
		UpdateColumn(t.Num, gorm.Expr("num - ?", 1))
	if err != nil {
		return false, err
	}
	return result.RowsAffected > 0, nil
}

func (r *TeamRepo) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.query.Team.WithContext(ctx).
		Where(r.query.Team.ID.Eq(id)).
		Delete()
	return err
}

func (r *TeamRepo) CreateWithCaptain(ctx context.Context, team *model.Team, captain *model.People) error {
	return r.query.Transaction(func(tx *query.Query) error {
		teamRepo := NewTeamRepoWithTx(tx)
		peopleRepo := NewPeopleRepoWithTx(tx)
		if err := teamRepo.Create(ctx, team); err != nil {
			return err
		}
		return peopleRepo.UpdateByID(ctx, captain.ID, map[string]any{
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
	err := r.query.Transaction(func(tx *query.Query) error {
		teamRepo := NewTeamRepoWithTx(tx)
		ok, err := teamRepo.IncrementNumIfAvailable(ctx, teamID, maxTeamSize)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if err := NewPeopleRepoWithTx(tx).UpdateByID(ctx, person.ID, updates); err != nil {
			return err
		}
		joined = true
		return nil
	})
	return joined, err
}

func (r *TeamRepo) RemoveMember(ctx context.Context, teamID int64, person *model.People) (bool, error) {
	removed := false
	err := r.query.Transaction(func(tx *query.Query) error {
		teamRepo := NewTeamRepoWithTx(tx)
		ok, err := teamRepo.DecrementNumIfPositive(ctx, teamID)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if err := NewPeopleRepoWithTx(tx).UpdateByID(ctx, person.ID, map[string]any{
			"role":    comm.RoleUnbind,
			"team_id": int64(-1),
		}); err != nil {
			return err
		}
		noticeRepo := NewNoticeRepoWithTx(tx)
		if err := noticeRepo.DeleteUnreadTypes(ctx, person.ID,
			comm.NoticeTeamPasswordChanged, comm.NoticeTeamRouteChanged); err != nil {
			return err
		}
		teamIDValue := teamID
		if err := noticeRepo.UpsertUnread(ctx, []*model.Notice{{
			UserID: person.ID, Type: comm.NoticeRemovedFromTeam, TeamID: &teamIDValue,
		}}); err != nil {
			return err
		}
		removed = true
		return nil
	})
	return removed, err
}

func (r *TeamRepo) ChangeCaptain(ctx context.Context, teamID, oldCaptainID, newCaptainID int64, oldCaptainName string) error {
	return r.query.Transaction(func(tx *query.Query) error {
		teamRepo := NewTeamRepoWithTx(tx)
		peopleRepo := NewPeopleRepoWithTx(tx)
		if err := teamRepo.UpdateByID(ctx, teamID, map[string]any{"captain": newCaptainID}); err != nil {
			return err
		}
		if err := peopleRepo.UpdateByID(ctx, oldCaptainID, map[string]any{"role": comm.RoleMember}); err != nil {
			return err
		}
		if err := peopleRepo.UpdateByID(ctx, newCaptainID, map[string]any{"role": comm.RoleCaptain}); err != nil {
			return err
		}
		teamIDValue, actorID := teamID, oldCaptainID
		return NewNoticeRepoWithTx(tx).UpsertUnread(ctx, []*model.Notice{{
			UserID: newCaptainID, Type: comm.NoticeCaptainTransferred,
			ActorID: &actorID, ActorName: oldCaptainName, TeamID: &teamIDValue,
		}})
	})
}

func (r *TeamRepo) DisbandTeam(ctx context.Context, teamID int64) error {
	return r.query.Transaction(func(tx *query.Query) error {
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
	t := r.query.Team
	var countRows []teamMatchCountRow
	if err := t.WithContext(ctx).
		Select(t.Num, t.ID.Count().As("count")).
		Where(t.RouteName.Eq(routeName), t.AllowMatch.Is(true), t.Num.Lt(uint8(maxTeamSize))).
		Group(t.Num).
		Scan(&countRows); err != nil {
		return nil, err
	}

	countsByNum := make(map[uint8]int64, len(countRows))
	for _, row := range countRows {
		countsByNum[row.Num] = row.Count
	}

	teams := make([]model.Team, 0)
	tiers := []struct {
		num   uint8
		count int64
		limit int
	}{
		{num: 3, count: countsByNum[1] + countsByNum[2] + countsByNum[3], limit: 3},
		{num: 4, count: countsByNum[4], limit: 4},
		{num: 5, count: countsByNum[5], limit: 5},
	}
	for _, tier := range tiers {
		if len(teams) >= 5 {
			break
		}
		take := tier.limit - len(teams)
		if take > int(tier.count) {
			take = int(tier.count)
		}
		if take <= 0 {
			continue
		}
		offset := randomOffset(tier.count, take)
		teamQuery := t.WithContext(ctx).Where(
			t.RouteName.Eq(routeName),
			t.AllowMatch.Is(true),
			t.Num.Lt(uint8(maxTeamSize)),
		)
		if tier.num == 3 {
			teamQuery = teamQuery.Where(t.Num.Lte(3))
		} else {
			teamQuery = teamQuery.Where(t.Num.Eq(tier.num))
		}
		rows, err := teamQuery.
			Order(t.ID).
			Offset(offset).
			Limit(take).
			Find()
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row != nil {
				teams = append(teams, *row)
			}
		}
	}
	return teams, nil
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

func (r *TeamRepo) FindRouteStartEdge(ctx context.Context, routeName string) (*model.RouteEdge, error) {
	re := r.query.RouteEdge
	record, err := re.WithContext(ctx).
		Where(
			re.RouteName.Eq(routeName),
			re.PrevPointName.IsNull(),
		).
		Order(re.SeqOrder).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
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

func (r *TeamRepo) IsPointOnRoute(ctx context.Context, routeName, pointName string) (bool, error) {
	re := r.query.RouteEdge
	total, err := re.WithContext(ctx).
		Where(re.RouteName.Eq(routeName), re.PointName.Eq(pointName)).
		Count()
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *TeamRepo) IsRouteTransitionValid(ctx context.Context, routeName, prevPointName, pointName string) (bool, error) {
	re := r.query.RouteEdge
	total, err := re.WithContext(ctx).
		Where(
			re.RouteName.Eq(routeName),
			re.PrevPointName.Eq(prevPointName),
			re.PointName.Eq(pointName),
		).
		Count()
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

	re := r.query.RouteEdge
	nextEdge, err := re.WithContext(ctx).
		Select(re.SeqOrder).
		Where(re.RouteName.Eq(routeName), re.PrevPointName.Eq(prevPointName)).
		Order(re.SeqOrder).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if nextEdge.SeqOrder == 0 {
		return false, nil
	}

	currentEdge, err := re.WithContext(ctx).
		Select(re.SeqOrder).
		Where(re.RouteName.Eq(routeName), re.PointName.Eq(pointName)).
		Order(re.SeqOrder.Desc()).
		First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if currentEdge.SeqOrder == 0 {
		return false, nil
	}
	return currentEdge.SeqOrder < nextEdge.SeqOrder, nil
}

func (r *TeamRepo) ListLatestCheckins(ctx context.Context, teamID int64, limit int) ([]TeamCheckinRow, error) {
	rows := make([]TeamCheckinRow, 0, limit)
	if limit <= 0 {
		return rows, nil
	}
	c := r.query.Checkin
	err := c.WithContext(ctx).
		Select(c.ID, c.AdminID, c.TeamID, c.PointName, c.RouteName, c.Time, c.CreatedAt).
		Where(c.TeamID.Eq(teamID)).
		Order(c.Time.Desc(), c.ID.Desc()).
		Limit(limit).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *TeamRepo) HasTeamCheckinAtPoint(ctx context.Context, teamID int64, pointName string) (bool, error) {
	c := r.query.Checkin
	total, err := c.WithContext(ctx).
		Where(c.TeamID.Eq(teamID), c.PointName.Eq(pointName)).
		Count()
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *TeamRepo) GetLatestWrongRouteName(ctx context.Context, teamID int64) (string, bool, error) {
	var row struct {
		WrongRouteName string `gorm:"column:wrong_route_name"`
	}
	w := r.query.WrongRouteRecord
	err := w.WithContext(ctx).
		Select(w.WrongRouteName).
		Where(w.TeamID.Eq(teamID)).
		Order(w.CreatedAt.Desc(), w.ID.Desc()).
		Limit(1).
		Scan(&row)
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
			ID:    row.ID,
			Name:  row.Name,
			Phone: row.Tel,
			Role:  row.Role,
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
	db := ndb.Pick().WithContext(ctx).
		Table("teams AS t").
		Joins("JOIN routes AS r ON r.name = t.route_name AND r.is_active = ? AND r.campus = ?", 1, query.Campus).
		Joins("LEFT JOIN peoples AS p ON p.team_id = t.id AND CAST(p.id AS CHAR) = t.captain").
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
		// （in_progress）。口径与 route/segment 的人数统计一致，避免
		// 尚未真正出发的队伍出现在路段队伍列表里。路段为空（仅关键词搜索）时不施加此过滤。
		db = db.Where(
			"EXISTS (SELECT 1 FROM peoples AS ps WHERE ps.team_id = t.id AND ps.walk_status = ?)",
			comm.WalkStatusInProgress,
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

	t := r.query.Team
	result, err := t.WithContext(ctx).
		Where(t.ID.Eq(teamID)).
		Updates(m)
	if err != nil {
		return false, err
	}

	return result.RowsAffected > 0, nil
}
