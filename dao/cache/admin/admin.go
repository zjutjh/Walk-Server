package admincache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"

	"app/dao/model"
)

const adminCacheTTL = time.Hour

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func adminKey(adminID int64) string {
	return "walk:admin:" + strconv.FormatInt(adminID, 10)
}

func GetAdmin(ctx context.Context, adminID int64) (*model.Admin, bool, error) {
	value, err := client().Get(ctx, adminKey(adminID)).Result()
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
	return client().Set(ctx, adminKey(admin.ID), payload, adminCacheTTL).Err()
}
