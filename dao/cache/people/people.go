package peoplecache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"

	"app/dao/model"
)

const (
	personByIDCacheKeyPrefix = "walk:user:profile"
	peopleCacheTTL           = time.Hour
)

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func BuildPersonByIDCacheKey(id int64) string {
	return personByIDCacheKeyPrefix + ":" + strconv.FormatInt(id, 10)
}

func SetPersonByID(ctx context.Context, people *model.People) error {
	if people == nil || people.ID <= 0 {
		return nil
	}

	payload, err := json.Marshal(people)
	if err != nil {
		return err
	}
	return client().Set(ctx, BuildPersonByIDCacheKey(people.ID), payload, peopleCacheTTL).Err()
}

func DelPersonByID(ctx context.Context, id int64) error {
	if id <= 0 {
		return nil
	}
	return client().Del(ctx, BuildPersonByIDCacheKey(id)).Err()
}
