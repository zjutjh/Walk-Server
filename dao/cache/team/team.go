package teamcache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"
)

const teamCacheTTL = time.Hour

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func teamIDByCodeKey(code string) string {
	return "walk:team_id_by_code:" + code
}

func GetTeamIDByCode(ctx context.Context, code string) (int64, bool, error) {
	value, err := client().Get(ctx, teamIDByCodeKey(code)).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	teamID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return teamID, true, nil
}

func SetTeamIDByCode(ctx context.Context, code string, teamID int64) error {
	return client().Set(ctx, teamIDByCodeKey(code), strconv.FormatInt(teamID, 10), teamCacheTTL).Err()
}
