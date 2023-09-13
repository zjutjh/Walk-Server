package global

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var (
	Rdb  *redis.Client
	Rctx = context.Background()
)
