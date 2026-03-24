package routecache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"

	"app/dao/model"
)

const routeCacheTTL = time.Hour

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func routeKey(routeName string) string {
	return "walk:route:" + routeName
}

func routeEdgeKey(routeName, pointName string) string {
	return "walk:route_edge:" + routeName + ":" + pointName
}

func pointRoutesKey(pointName string) string {
	return "walk:point:routes:" + pointName
}

func GetRoute(ctx context.Context, routeName string) (*model.Route, bool, error) {
	value, err := client().Get(ctx, routeKey(routeName)).Result()
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
	return client().Set(ctx, routeKey(route.Name), payload, routeCacheTTL).Err()
}

func GetRouteEdge(ctx context.Context, routeName, pointName string) (*model.RouteEdge, bool, error) {
	value, err := client().Get(ctx, routeEdgeKey(routeName, pointName)).Result()
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
	return client().Set(ctx, routeEdgeKey(routeEdge.RouteName, routeEdge.PointName), payload, routeCacheTTL).Err()
}

func GetPointRoutes(ctx context.Context, pointName string) ([]string, bool, error) {
	value, err := client().Get(ctx, pointRoutesKey(pointName)).Result()
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
	return client().Set(ctx, pointRoutesKey(pointName), payload, routeCacheTTL).Err()
}
