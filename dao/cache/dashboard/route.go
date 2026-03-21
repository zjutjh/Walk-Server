package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
)

type DashboardCache struct {
	rdb redis.UniversalClient
}

const (
	allRouteStatsCacheKey    = "dashboard:stats:route:all"
	allRouteStatsCacheTTL    = 15 * time.Second
	segmentCacheKeyPrefix    = "dashboard:segment"
	segmentCacheTTL          = 15 * time.Second
	checkpointCacheKeyPrefix = "dashboard:checkpoint"
	checkpointCacheTTL       = 15 * time.Second

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

func BuildSegmentCacheKey(campus string, prevPointName string, toPointName string) string {
	return fmt.Sprintf("%s:%s:%s:%s", segmentCacheKeyPrefix, campus, prevPointName, toPointName)
}

func BuildCheckpointCacheKey(campus string, pointName string) string {
	return fmt.Sprintf("%s:%s:%s", checkpointCacheKeyPrefix, campus, pointName)
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

// GetSegment 返回缓存内容；命中返回 found=true，未命中返回 found=false。
func (c *DashboardCache) GetSegment(ctx context.Context, campus string, prevPointName string, toPointName string) (cached []byte, found bool, err error) {
	cached, err = c.rdb.Get(ctx, BuildSegmentCacheKey(campus, prevPointName, toPointName)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return cached, true, nil
}

func (c *DashboardCache) SetSegment(ctx context.Context, campus string, prevPointName string, toPointName string, cached []byte) error {
	return c.rdb.Set(ctx, BuildSegmentCacheKey(campus, prevPointName, toPointName), cached, segmentCacheTTL).Err()
}

// GetCheckpoint 返回缓存内容；命中返回 found=true，未命中返回 found=false。
func (c *DashboardCache) GetCheckpoint(ctx context.Context, campus string, pointName string) (cached []byte, found bool, err error) {
	cached, err = c.rdb.Get(ctx, BuildCheckpointCacheKey(campus, pointName)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return cached, true, nil
}

func (c *DashboardCache) SetCheckpoint(ctx context.Context, campus string, pointName string, cached []byte) error {
	return c.rdb.Set(ctx, BuildCheckpointCacheKey(campus, pointName), cached, checkpointCacheTTL).Err()
}
