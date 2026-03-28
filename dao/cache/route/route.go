package routeCache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"

	"app/dao/model"
)

const (
	routeCacheKeyPrefix       = "walk:route"
	routeEdgeCacheKeyPrefix   = "walk:route_edge"
	pointRoutesCacheKeyPrefix = "walk:point:routes"
	routeCacheTTL             = time.Hour

	allRouteStatsCacheKey         = "dashboard:stats:route:all"
	allRouteStatsCacheTTL         = 15 * time.Second
	overviewCacheKeyPrefix        = "dashboard:overview"
	overviewCacheTTL              = 15 * time.Second
	segmentCacheKeyPrefix         = "dashboard:segment"
	segmentCacheTTL               = 15 * time.Second
	checkpointCacheKeyPrefix      = "dashboard:checkpoint"
	checkpointCacheTTL            = 15 * time.Second
	routeDetailStatsCacheKeyPrefix = "dashboard:stats:route:detail"
	routeDetailStatsCacheTTL      = 15 * time.Second
)

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func BuildRouteCacheKey(routeName string) string {
	return fmt.Sprintf("%s:%s", routeCacheKeyPrefix, routeName)
}

func BuildRouteEdgeCacheKey(routeName, pointName string) string {
	return fmt.Sprintf("%s:%s:%s", routeEdgeCacheKeyPrefix, routeName, pointName)
}

func BuildPointRoutesCacheKey(pointName string) string {
	return fmt.Sprintf("%s:%s", pointRoutesCacheKeyPrefix, pointName)
}

func BuildRouteDetailStatsCacheKey(routeName string) string {
	return fmt.Sprintf("%s:%s", routeDetailStatsCacheKeyPrefix, routeName)
}

func BuildOverviewCacheKey(campus string) string {
	return fmt.Sprintf("%s:%s", overviewCacheKeyPrefix, campus)
}

func BuildSegmentCacheKey(campus, prevPointName, toPointName string) string {
	return fmt.Sprintf("%s:%s:%s:%s", segmentCacheKeyPrefix, campus, prevPointName, toPointName)
}

func BuildCheckpointCacheKey(campus, pointName string) string {
	return fmt.Sprintf("%s:%s:%s", checkpointCacheKeyPrefix, campus, pointName)
}

func GetRoute(ctx context.Context, routeName string) (*model.Route, bool, error) {
	value, err := client().Get(ctx, BuildRouteCacheKey(routeName)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var route model.Route
	if err := json.Unmarshal([]byte(value), &route); err != nil {
		return nil, false, err
	}
	return &route, true, nil
}

func SetRoute(ctx context.Context, route *model.Route) error {
	if route == nil {
		return nil
	}

	payload, err := json.Marshal(route)
	if err != nil {
		return err
	}
	return client().Set(ctx, BuildRouteCacheKey(route.Name), payload, routeCacheTTL).Err()
}

func GetRouteEdge(ctx context.Context, routeName, pointName string) (*model.RouteEdge, bool, error) {
	value, err := client().Get(ctx, BuildRouteEdgeCacheKey(routeName, pointName)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var routeEdge model.RouteEdge
	if err := json.Unmarshal([]byte(value), &routeEdge); err != nil {
		return nil, false, err
	}
	return &routeEdge, true, nil
}

func SetRouteEdge(ctx context.Context, routeEdge *model.RouteEdge) error {
	if routeEdge == nil {
		return nil
	}

	payload, err := json.Marshal(routeEdge)
	if err != nil {
		return err
	}
	return client().Set(ctx, BuildRouteEdgeCacheKey(routeEdge.RouteName, routeEdge.PointName), payload, routeCacheTTL).Err()
}

func GetPointRoutes(ctx context.Context, pointName string) ([]string, bool, error) {
	value, err := client().Get(ctx, BuildPointRoutesCacheKey(pointName)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var routeNames []string
	if err := json.Unmarshal([]byte(value), &routeNames); err != nil {
		return nil, false, err
	}
	return routeNames, true, nil
}

func SetPointRoutes(ctx context.Context, pointName string, routeNames []string) error {
	payload, err := json.Marshal(routeNames)
	if err != nil {
		return err
	}
	return client().Set(ctx, BuildPointRoutesCacheKey(pointName), payload, routeCacheTTL).Err()
}

func GetAllRouteStats(ctx context.Context) ([]byte, bool, error) {
	cached, err := client().Get(ctx, allRouteStatsCacheKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cached, true, nil
}

func SetAllRouteStats(ctx context.Context, cached []byte) error {
	return client().Set(ctx, allRouteStatsCacheKey, cached, allRouteStatsCacheTTL).Err()
}

func GetOverview(ctx context.Context, campus string) ([]byte, bool, error) {
	cached, err := client().Get(ctx, BuildOverviewCacheKey(campus)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cached, true, nil
}

func SetOverview(ctx context.Context, campus string, cached []byte) error {
	return client().Set(ctx, BuildOverviewCacheKey(campus), cached, overviewCacheTTL).Err()
}

func GetRouteDetailStats(ctx context.Context, routeName string) ([]byte, bool, error) {
	cached, err := client().Get(ctx, BuildRouteDetailStatsCacheKey(routeName)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cached, true, nil
}

func SetRouteDetailStats(ctx context.Context, routeName string, cached []byte) error {
	return client().Set(ctx, BuildRouteDetailStatsCacheKey(routeName), cached, routeDetailStatsCacheTTL).Err()
}

func GetSegment(ctx context.Context, campus, prevPointName, toPointName string) ([]byte, bool, error) {
	cached, err := client().Get(ctx, BuildSegmentCacheKey(campus, prevPointName, toPointName)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cached, true, nil
}

func SetSegment(ctx context.Context, campus, prevPointName, toPointName string, cached []byte) error {
	return client().Set(ctx, BuildSegmentCacheKey(campus, prevPointName, toPointName), cached, segmentCacheTTL).Err()
}

func GetCheckpoint(ctx context.Context, campus, pointName string) ([]byte, bool, error) {
	cached, err := client().Get(ctx, BuildCheckpointCacheKey(campus, pointName)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cached, true, nil
}

func SetCheckpoint(ctx context.Context, campus, pointName string, cached []byte) error {
	return client().Set(ctx, BuildCheckpointCacheKey(campus, pointName), cached, checkpointCacheTTL).Err()
}
