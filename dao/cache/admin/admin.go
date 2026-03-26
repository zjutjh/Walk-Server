package admincache

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
	adminCacheKeyPrefix = "walk:admin"
	adminCacheTTL       = time.Hour
)

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func BuildAdminCacheKey(adminID int64) string {
	return fmt.Sprintf("%s:%d", adminCacheKeyPrefix, adminID)
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
