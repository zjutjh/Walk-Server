package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"

	"app/dao/model"
)

const (
	adminCacheKeyPrefix          = "walk:admin"
	adminLoginFailCacheKeyPrefix = "walk:admin:login_fail"
	adminCacheTTL                = time.Hour
	adminLoginFailTTL            = 10 * time.Minute
	adminLoginFailMaxCount       = 5
)

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func BuildAdminCacheKey(adminID int64) string {
	return fmt.Sprintf("%s:%d", adminCacheKeyPrefix, adminID)
}

func BuildAdminLoginFailCacheKey(account string) string {
	return fmt.Sprintf("%s:%s", adminLoginFailCacheKeyPrefix, account)
}

func GetAdmin(ctx context.Context, adminID int64) (*model.Admin, bool, error) {
	value, err := client().Get(ctx, BuildAdminCacheKey(adminID)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var admin model.Admin
	if err := json.Unmarshal([]byte(value), &admin); err != nil {
		return nil, false, err
	}
	return &admin, true, nil
}

func SetAdmin(ctx context.Context, admin *model.Admin) error {
	if admin == nil {
		return nil
	}

	payload, err := json.Marshal(admin)
	if err != nil {
		return err
	}
	return client().Set(ctx, BuildAdminCacheKey(admin.ID), payload, adminCacheTTL).Err()
}

func IsAdminLoginBlocked(ctx context.Context, account string) (bool, error) {
	if account == "" {
		return false, nil
	}

	count, err := client().Get(ctx, BuildAdminLoginFailCacheKey(account)).Int()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return count >= adminLoginFailMaxCount, nil
}

func IncrementAdminLoginFail(ctx context.Context, account string) error {
	if account == "" {
		return nil
	}

	key := BuildAdminLoginFailCacheKey(account)
	count, err := client().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		return client().Expire(ctx, key, adminLoginFailTTL).Err()
	}
	return nil
}

func ClearAdminLoginFail(ctx context.Context, account string) error {
	if account == "" {
		return nil
	}
	return client().Del(ctx, BuildAdminLoginFailCacheKey(account)).Err()
}
