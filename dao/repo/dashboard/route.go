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
		Table("people AS p").
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
		Table("people AS p").
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
