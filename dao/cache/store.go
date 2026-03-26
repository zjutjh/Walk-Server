package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
)

func GetJSON[T any](ctx context.Context, key string) (T, bool, error) {
	var zero T

	raw, err := nedis.Pick().Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return zero, false, nil
	}
	if err != nil {
		return zero, false, err
	}

	var value T
	if err = json.Unmarshal([]byte(raw), &value); err != nil {
		return zero, false, err
	}

	return value, true, nil
}

func SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return nedis.Pick().Set(ctx, key, payload, ttl).Err()
}

func Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	return nedis.Pick().Del(ctx, keys...).Err()
}
