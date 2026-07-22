package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
	"github.com/zjutjh/mygo/lock"
	"github.com/zjutjh/mygo/nedis"

	"app/dao/model"
)

const (
	teamIDByCodeCacheKeyPrefix = "walk:team_id_by_code"
	teamByIDCacheKeyPrefix     = "walk:user:team:info"
	teamCacheTTL               = time.Hour
	submittedTeamsKey          = "teams"
	teamInfoCacheKeyPrefix     = "dashboard:teams:info"
	teamInfoCacheTTL           = 60 * time.Second
	teamFilterCacheKeyPrefix   = "dashboard:teams:filter"
	teamFilterCacheTTL         = 30 * time.Second
	teamInfoLockCacheKeyPrefix = "dashboard:teams:info:lock"
)

var teamInfoLocks sync.Map

var submitTeamScript = redis.NewScript(`
local teamID = KEYS[1]
local dailyRouteKey = KEYS[2]

local num = redis.call("GET", dailyRouteKey)
if not num or tonumber(num) <= 0 then
	return 2
end

local added = redis.call("SADD", "teams", teamID)
if added == 0 then
	return 1
end

redis.call("DECR", dailyRouteKey)
return 0
`)

func client() redis.UniversalClient {
	return nedis.Pick("redis")
}

func BuildTeamIDByCodeCacheKey(code string) string {
	return fmt.Sprintf("%s:%s", teamIDByCodeCacheKeyPrefix, code)
}

func BuildTeamByIDCacheKey(teamID int64) string {
	return fmt.Sprintf("%s:%d", teamByIDCacheKeyPrefix, teamID)
}

func BuildTeamInfoCacheKey(teamID int64) string {
	return fmt.Sprintf("%s:%d", teamInfoCacheKeyPrefix, teamID)
}

func BuildTeamFilterCacheKey(campus, queryHash string) string {
	return fmt.Sprintf("%s:%s:%s", teamFilterCacheKeyPrefix, campus, queryHash)
}

func BuildTeamInfoLockCacheKey(teamID int64) string {
	return fmt.Sprintf("%s:%d", teamInfoLockCacheKeyPrefix, teamID)
}

func buildDailyRouteQuotaKey(day int, routeCode int) string {
	return strconv.Itoa(day*10 + routeCode)
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

func DelTeamIDByCode(ctx context.Context, code string) error {
	if code == "" {
		return nil
	}
	return client().Del(ctx, BuildTeamIDByCodeCacheKey(code)).Err()
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

func IsTeamSubmitted(ctx context.Context, teamID int64) (bool, error) {
	return client().SIsMember(ctx, submittedTeamsKey, strconv.FormatInt(teamID, 10)).Result()
}

func SubmitTeam(ctx context.Context, teamID int64, day int, routeCode int) (int64, error) {
	return submitTeamScript.Run(
		ctx,
		client(),
		[]string{
			strconv.FormatInt(teamID, 10),
			buildDailyRouteQuotaKey(day, routeCode),
		},
	).Int64()
}

func RollbackTeamSubmit(ctx context.Context, teamID int64, day int, routeCode int) (bool, error) {
	teamIDValue := strconv.FormatInt(teamID, 10)
	submitted, err := client().SIsMember(ctx, submittedTeamsKey, teamIDValue).Result()
	if err != nil {
		return false, err
	}
	if !submitted {
		return false, nil
	}
	if err := client().SRem(ctx, submittedTeamsKey, teamIDValue).Err(); err != nil {
		return false, err
	}
	if err := client().Incr(ctx, buildDailyRouteQuotaKey(day, routeCode)).Err(); err != nil {
		return false, err
	}
	return true, nil
}

func RestoreSubmittedTeam(ctx context.Context, teamID int64, day int, routeCode int) error {
	if err := client().SAdd(ctx, submittedTeamsKey, strconv.FormatInt(teamID, 10)).Err(); err != nil {
		return err
	}
	return client().Decr(ctx, buildDailyRouteQuotaKey(day, routeCode)).Err()
}

func InitDailyRouteQuota(ctx context.Context, day int, routeCode int, limit int) error {
	key := buildDailyRouteQuotaKey(day, routeCode)
	if _, err := client().Get(ctx, key).Result(); err == redis.Nil {
		return client().Set(ctx, key, limit, 0).Err()
	} else if err != nil {
		return err
	}
	return nil
}

func GetTeamInfo(ctx context.Context, teamID int64) ([]byte, bool, error) {
	cached, err := client().Get(ctx, BuildTeamInfoCacheKey(teamID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cached, true, nil
}

func SetTeamInfo(ctx context.Context, teamID int64, cached []byte) error {
	return client().Set(ctx, BuildTeamInfoCacheKey(teamID), cached, teamInfoCacheTTL).Err()
}

func DeleteTeamInfo(ctx context.Context, teamID int64) error {
	return client().Del(ctx, BuildTeamInfoCacheKey(teamID)).Err()
}

func GetTeamFilter(ctx context.Context, campus, queryHash string) ([]byte, bool, error) {
	cached, err := client().Get(ctx, BuildTeamFilterCacheKey(campus, queryHash)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return cached, true, nil
}

func SetTeamFilter(ctx context.Context, campus, queryHash string, cached []byte) error {
	return client().Set(ctx, BuildTeamFilterCacheKey(campus, queryHash), cached, teamFilterCacheTTL).Err()
}

func getTeamInfoMutex(teamID int64) (*redsync.Mutex, bool) {
	value, ok := teamInfoLocks.Load(teamID)
	if !ok {
		return nil, false
	}

	mutex, ok := value.(*redsync.Mutex)
	if !ok || mutex == nil {
		teamInfoLocks.Delete(teamID)
		return nil, false
	}

	return mutex, true
}

func setTeamInfoMutex(teamID int64, mutex *redsync.Mutex) {
	if mutex == nil {
		teamInfoLocks.Delete(teamID)
		return
	}
	teamInfoLocks.Store(teamID, mutex)
}

func AcquireTeamInfoLock(ctx context.Context, teamID int64, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, nil
	}

	mutex := lock.Pick().NewMutex(
		BuildTeamInfoLockCacheKey(teamID),
		redsync.WithExpiry(ttl),
		redsync.WithTries(1),
	)

	err := mutex.LockContext(ctx)
	if errors.Is(err, redsync.ErrFailed) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	setTeamInfoMutex(teamID, mutex)
	return true, nil
}

func SetTeamInfoLockTTL(ctx context.Context, teamID int64, ttl time.Duration) error {
	if ttl <= 0 {
		return ReleaseTeamInfoLock(ctx, teamID)
	}

	current, ok := getTeamInfoMutex(teamID)
	if !ok {
		return nil
	}

	mutex := lock.Pick().NewMutex(
		BuildTeamInfoLockCacheKey(teamID),
		redsync.WithExpiry(ttl),
		redsync.WithTries(1),
		redsync.WithValue(current.Value()),
	)

	extended, err := mutex.ExtendContext(ctx)
	if err != nil {
		return err
	}
	if !extended {
		return nil
	}

	setTeamInfoMutex(teamID, mutex)
	return nil
}

func ReleaseTeamInfoLock(ctx context.Context, teamID int64) error {
	mutex, ok := getTeamInfoMutex(teamID)
	if !ok {
		return nil
	}

	defer teamInfoLocks.Delete(teamID)

	unlocked, err := mutex.UnlockContext(ctx)
	if err != nil {
		return err
	}
	if !unlocked {
		return nil
	}
	return nil
}
