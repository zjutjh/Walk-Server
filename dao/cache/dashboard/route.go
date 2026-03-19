package cache

import (
	"context"
	"errors"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
)

type DashboardCache struct {
	rdb redis.UniversalClient
}

const (
	allRouteStatsCacheKey = "dashboard:stats:route:all"
	allRouteStatsCacheTTL = 15 * time.Second

	// routeDetailStatsCacheKeyPrefix 预留给 /dashboard/stats/route 单路线统计接口。
	// 实际 key 形如: dashboard:stats:route:detail:pf-half
	routeDetailStatsCacheKeyPrefix = "dashboard:stats:route:detail"
)

func NewDashboardCache() *DashboardCache {
	return &DashboardCache{rdb: pickStatsRedis()}
}

func pickStatsRedis() redis.UniversalClient {
	if nedis.Exist("cache") {
		return nedis.Pick("cache")
	}

	return nedis.Pick()
}

func BuildRouteDetailStatsCacheKey(routeName string) string {
	return routeDetailStatsCacheKeyPrefix + ":" + routeName
}

// GetAllRouteStats 返回缓存内容；命中返回 found=true，未命中返回 found=false。
func (c *DashboardCache) GetAllRouteStats(ctx context.Context) (cached []byte, found bool, err error) {
	cached, err = c.rdb.Get(ctx, allRouteStatsCacheKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return cached, true, nil
}

func (c *DashboardCache) SetAllRouteStats(ctx context.Context, cached []byte) error {
	return c.rdb.Set(ctx, allRouteStatsCacheKey, cached, allRouteStatsCacheTTL).Err()
}
