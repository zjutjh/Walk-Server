package teamCache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/nedis"

	baseCache "app/dao/cache"
	"app/dao/model"
)

const (
	teamIDByCodeCacheKeyPrefix = "walk:team_id_by_code"
	teamCacheTTL               = time.Hour
)

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func BuildTeamIDByCodeCacheKey(code string) string {
	return fmt.Sprintf("%s:%s", teamIDByCodeCacheKeyPrefix, code)
}

func BuildTeamByIDCacheKey(teamID int64) string {
	return baseCache.TeamInfoKey(teamID)
}

func GetTeamIDByCode(ctx context.Context, code string) (int64, bool, error) {
	value, err := client().Get(ctx, BuildTeamIDByCodeCacheKey(code)).Result()
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
	return client().Set(ctx, BuildTeamIDByCodeCacheKey(code), strconv.FormatInt(teamID, 10), teamCacheTTL).Err()
}

func GetTeamByID(ctx context.Context, teamID int64) (*model.Team, bool, error) {
	value, err := client().Get(ctx, BuildTeamByIDCacheKey(teamID)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var team model.Team
	if err := json.Unmarshal([]byte(value), &team); err != nil {
		return nil, false, err
	}
	return &team, true, nil
}

func SetTeamByID(ctx context.Context, team *model.Team) error {
	if team == nil {
		return nil
	}

	payload, err := json.Marshal(team)
	if err != nil {
		return err
	}
	return client().Set(ctx, BuildTeamByIDCacheKey(team.ID), payload, teamCacheTTL).Err()
}

func DelTeamByID(ctx context.Context, teamID int64) error {
	if teamID <= 0 {
		return nil
	}
	return client().Del(ctx, BuildTeamByIDCacheKey(teamID)).Err()
}
