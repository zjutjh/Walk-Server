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
	teamIDByCodeCacheKeyPrefix = "walk:team:by-code"
	teamByIDCacheKeyPrefix     = "walk:team:by-id"
	teamCacheTTL               = time.Hour
	submittedTeamsKey          = "walk:team:submitted"
	submittedTeamDaysKey       = "walk:team:submitted-day"
	submittedTeamRoutesKey     = "walk:team:submitted-route"
	totalTeamQuotaKey          = "walk:team:quota:total"
	dailyTeamQuotaKeyPrefix    = "walk:team:quota:route:"
	teamInfoCacheKeyPrefix     = "walk:dashboard:team:by-id"
	teamInfoCacheTTL           = 60 * time.Second
	teamFilterCacheKeyPrefix   = "walk:dashboard:team:filter"
	teamFilterCacheTTL         = 30 * time.Second
	teamInfoLockCacheKeyPrefix = "walk:lock:dashboard:team"
)

var teamInfoLocks sync.Map

var submitTeamScript = redis.NewScript(`
local submittedTeamsKey = KEYS[1]
local dailyQuotaKey = KEYS[2]
local totalQuotaKey = KEYS[3]
local submittedDaysKey = KEYS[4]
local submittedRoutesKey = KEYS[5]
local teamID = ARGV[1]
local day = ARGV[2]
local routeName = ARGV[3]

local submitted = redis.call("SISMEMBER", submittedTeamsKey, teamID)
if submitted == 1 then
	return 1
end

local total = redis.call("GET", totalQuotaKey)
if not total or tonumber(total) <= 0 then
	return 3
end

local daily = redis.call("GET", dailyQuotaKey)
if not daily or tonumber(daily) <= 0 then
	return 2
end

redis.call("SADD", submittedTeamsKey, teamID)
redis.call("HSET", submittedDaysKey, teamID, day)
redis.call("HSET", submittedRoutesKey, teamID, routeName)
redis.call("DECR", dailyQuotaKey)
redis.call("DECR", totalQuotaKey)
return 0
`)

var rollbackTeamSubmitScript = redis.NewScript(`
local submittedTeamsKey = KEYS[1]
local submittedDaysKey = KEYS[2]
local totalQuotaKey = KEYS[3]
local submittedRoutesKey = KEYS[4]
local teamID = ARGV[1]
local fallbackDay = ARGV[2]
local fallbackRoute = ARGV[3]
local dailyQuotaKeyPrefix = ARGV[4]

if redis.call("SISMEMBER", submittedTeamsKey, teamID) == 0 then
	return {0, "", 0}
end

local day = redis.call("HGET", submittedDaysKey, teamID)
if not day then
	if tonumber(fallbackDay) < 0 then
		return redis.error_reply("submitted team day is missing")
	end
	day = fallbackDay
end

local routeName = redis.call("HGET", submittedRoutesKey, teamID)
if not routeName then
	if fallbackRoute == "" then
		return redis.error_reply("submitted team route is missing")
	end
	routeName = fallbackRoute
end

redis.call("SREM", submittedTeamsKey, teamID)
redis.call("HDEL", submittedDaysKey, teamID)
redis.call("HDEL", submittedRoutesKey, teamID)
redis.call("INCR", dailyQuotaKeyPrefix .. routeName .. ":day:" .. day)
redis.call("INCR", totalQuotaKey)
return {1, routeName, tonumber(day)}
`)

var restoreSubmittedTeamScript = redis.NewScript(`
local submittedTeamsKey = KEYS[1]
local submittedDaysKey = KEYS[2]
local dailyQuotaKey = KEYS[3]
local totalQuotaKey = KEYS[4]
local submittedRoutesKey = KEYS[5]
local teamID = ARGV[1]
local day = ARGV[2]
local routeName = ARGV[3]

if redis.call("SISMEMBER", submittedTeamsKey, teamID) == 1 then
	return 0
end

redis.call("SADD", submittedTeamsKey, teamID)
redis.call("HSET", submittedDaysKey, teamID, day)
redis.call("HSET", submittedRoutesKey, teamID, routeName)
redis.call("DECR", dailyQuotaKey)
redis.call("DECR", totalQuotaKey)
return 1
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

func buildDailyTeamQuotaKey(routeName string, day int) string {
	return dailyTeamQuotaKeyPrefix + routeName + ":day:" + strconv.Itoa(day)
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

func SubmitTeam(ctx context.Context, teamID int64, routeName string, day int) (int64, error) {
	return submitTeamScript.Run(
		ctx,
		client(),
		[]string{
			submittedTeamsKey,
			buildDailyTeamQuotaKey(routeName, day),
			totalTeamQuotaKey,
			submittedTeamDaysKey,
			submittedTeamRoutesKey,
		},
		teamID,
		day,
		routeName,
	).Int64()
}

// RollbackTeamSubmit refunds the original submission day. A negative fallbackDay
// requires the original day to exist, instead of guessing which day's quota to refund.
func RollbackTeamSubmit(ctx context.Context, teamID int64, fallbackRoute string, fallbackDay int) (bool, string, int, error) {
	teamIDValue := strconv.FormatInt(teamID, 10)
	result, err := rollbackTeamSubmitScript.Run(
		ctx,
		client(),
		[]string{submittedTeamsKey, submittedTeamDaysKey, totalTeamQuotaKey, submittedTeamRoutesKey},
		teamIDValue,
		fallbackDay,
		fallbackRoute,
		dailyTeamQuotaKeyPrefix,
	).Slice()
	if err != nil {
		return false, "", 0, err
	}
	if len(result) != 3 {
		return false, "", 0, fmt.Errorf("unexpected rollback team submit result: %v", result)
	}
	removed, ok := result[0].(int64)
	if !ok {
		return false, "", 0, fmt.Errorf("unexpected rollback flag: %v", result[0])
	}
	if removed == 0 {
		return false, "", 0, nil
	}
	routeName, ok := result[1].(string)
	day, dayOK := result[2].(int64)
	if !ok || !dayOK {
		return false, "", 0, fmt.Errorf("unexpected rollback metadata: %v", result)
	}
	return true, routeName, int(day), nil
}

func RestoreSubmittedTeam(ctx context.Context, teamID int64, routeName string, day int) error {
	teamIDValue := strconv.FormatInt(teamID, 10)
	return restoreSubmittedTeamScript.Run(
		ctx,
		client(),
		[]string{
			submittedTeamsKey,
			submittedTeamDaysKey,
			buildDailyTeamQuotaKey(routeName, day),
			totalTeamQuotaKey,
			submittedTeamRoutesKey,
		},
		teamIDValue,
		day,
		routeName,
	).Err()
}

// ClearSubmittedTeam removes the submission ledger without refunding quota.
// Used after a team is disbanded during adjustment, when registration is closed.
func ClearSubmittedTeam(ctx context.Context, teamID int64) error {
	return clearSubmittedTeam(ctx, client(), teamID)
}

func clearSubmittedTeam(ctx context.Context, redisClient redis.UniversalClient, teamID int64) error {
	_, err := redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, submittedTeamsKey, teamID)
		pipe.HDel(ctx, submittedTeamDaysKey, strconv.FormatInt(teamID, 10))
		pipe.HDel(ctx, submittedTeamRoutesKey, strconv.FormatInt(teamID, 10))
		return nil
	})
	return err
}

func InitDailyTeamQuota(ctx context.Context, routeName string, day int, limit int) error {
	key := buildDailyTeamQuotaKey(routeName, day)
	if _, err := client().Get(ctx, key).Result(); err == redis.Nil {
		return client().Set(ctx, key, limit, 0).Err()
	} else if err != nil {
		return err
	}
	return nil
}

func InitTotalTeamQuota(ctx context.Context, limit int) error {
	return client().SetNX(ctx, totalTeamQuotaKey, limit, 0).Err()
}

// GetDailyTeamQuotaAvailability reports whether the global quota and each
// route's quota are both available for the specified submission day.
func GetDailyTeamQuotaAvailability(ctx context.Context, routeNames []string, day int) (map[string]bool, error) {
	return getDailyTeamQuotaAvailability(ctx, client(), routeNames, day)
}

func getDailyTeamQuotaAvailability(ctx context.Context, redisClient redis.UniversalClient, routeNames []string, day int) (map[string]bool, error) {
	keys := make([]string, 1, len(routeNames)+1)
	keys[0] = totalTeamQuotaKey
	for _, routeName := range routeNames {
		keys = append(keys, buildDailyTeamQuotaKey(routeName, day))
	}
	values, err := redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	if len(values) != len(keys) {
		return nil, fmt.Errorf("unexpected team quota result length: %d", len(values))
	}
	parseQuota := func(key string, value any) (int64, error) {
		if value == nil {
			return 0, fmt.Errorf("team quota key %s is missing", key)
		}
		quota, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse team quota key %s: %w", key, err)
		}
		return quota, nil
	}
	total, err := parseQuota(keys[0], values[0])
	if err != nil {
		return nil, err
	}
	availability := make(map[string]bool, len(routeNames))
	for index, routeName := range routeNames {
		routeQuota, err := parseQuota(keys[index+1], values[index+1])
		if err != nil {
			return nil, err
		}
		availability[routeName] = total > 0 && routeQuota > 0
	}
	return availability, nil
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
