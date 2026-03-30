package peopleCache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"

	baseCache "app/dao/cache"
	"app/dao/model"
)

const (
	peopleCacheTTL = time.Hour
)

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func BuildPersonByOpenIDCacheKey(openID string) string {
	return baseCache.UserProfileKey(openID)
}

func GetPersonByOpenID(ctx context.Context, openID string) (*model.People, bool, error) {
	value, err := client().Get(ctx, BuildPersonByOpenIDCacheKey(openID)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var people model.People
	if err := json.Unmarshal([]byte(value), &people); err != nil {
		return nil, false, err
	}
	return &people, true, nil
}

func SetPersonByOpenID(ctx context.Context, people *model.People) error {
	if people == nil || people.OpenID == "" {
		return nil
	}

	payload, err := json.Marshal(people)
	if err != nil {
		return err
	}
	return client().Set(ctx, BuildPersonByOpenIDCacheKey(people.OpenID), payload, peopleCacheTTL).Err()
}

func DelPersonByOpenID(ctx context.Context, openID string) error {
	if openID == "" {
		return nil
	}
	return client().Del(ctx, BuildPersonByOpenIDCacheKey(openID)).Err()
}
