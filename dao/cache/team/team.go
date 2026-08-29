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
	totalTeamQuotaKey          = "walk:team:quota:total"
	dailyTeamQuotaKeyPrefix    = "walk:team:quota:day:"
	teamInfoCacheKeyPrefix     = "walk:dashboard:team:by-id"
	teamInfoCacheTTL           = 60 * time.Second
	teamFilterCacheKeyPrefix   = "walk:dashboard:team:filter"
	teamFilterCacheTTL         = 30 * time.Second
	teamInfoLockCacheKeyPrefix = "walk:lock:dashboard:team"
	teamChangeNoticeKeyPrefix  = "walk:team:change-notice"
	teamChangeNoticeTTL        = 30 * 24 * time.Hour
)

var teamInfoLocks sync.Map

var submitTeamScript = redis.NewScript(`
local submittedTeamsKey = KEYS[1]
local dailyQuotaKey = KEYS[2]
local totalQuotaKey = KEYS[3]
local submittedDaysKey = KEYS[4]
local teamID = ARGV[1]
local day = ARGV[2]

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
redis.call("DECR", dailyQuotaKey)
redis.call("DECR", totalQuotaKey)
return 0
`)

var rollbackTeamSubmitScript = redis.NewScript(`
local submittedTeamsKey = KEYS[1]
local submittedDaysKey = KEYS[2]
local totalQuotaKey = KEYS[3]
local teamID = ARGV[1]
local fallbackDay = ARGV[2]
local dailyQuotaKeyPrefix = ARGV[3]

if redis.call("SISMEMBER", submittedTeamsKey, teamID) == 0 then
	return {0, 0}
end

local day = redis.call("HGET", submittedDaysKey, teamID)
if not day then
	day = fallbackDay
end

redis.call("SREM", submittedTeamsKey, teamID)
redis.call("HDEL", submittedDaysKey, teamID)
redis.call("INCR", dailyQuotaKeyPrefix .. day)
redis.call("INCR", totalQuotaKey)
return {1, tonumber(day)}
`)

var restoreSubmittedTeamScript = redis.NewScript(`
local submittedTeamsKey = KEYS[1]
local submittedDaysKey = KEYS[2]
local dailyQuotaKey = KEYS[3]
local totalQuotaKey = KEYS[4]
local teamID = ARGV[1]
local day = ARGV[2]

if redis.call("SISMEMBER", submittedTeamsKey, teamID) == 1 then
	return 0
end

redis.call("SADD", submittedTeamsKey, teamID)
redis.call("HSET", submittedDaysKey, teamID, day)
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

func buildTeamChangeNoticeKey(userID int64) string {
	return fmt.Sprintf("%s:%d", teamChangeNoticeKeyPrefix, userID)
}

// SetTeamChangeNotice 为队员记录尚未查看的团队密码、路线变更通知。
func SetTeamChangeNotice(ctx context.Context, userIDs []int64, passwordChanged, routeChanged bool) error {
	if len(userIDs) == 0 || (!passwordChanged && !routeChanged) {
		return nil
	}

	pipe := client().Pipeline()
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		key := buildTeamChangeNoticeKey(userID)
		values := make(map[string]any, 2)
		if passwordChanged {
			values["password_changed"] = 1
		}
		if routeChanged {
			values["route_changed"] = 1
		}
		pipe.HSet(ctx, key, values)
		pipe.Expire(ctx, key, teamChangeNoticeTTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// GetTeamChangeNotice 读取用户尚未确认的团队变更通知，不改变已读状态。
func GetTeamChangeNotice(ctx context.Context, userID int64) (bool, bool, error) {
	result, err := client().HGetAll(ctx, buildTeamChangeNoticeKey(userID)).Result()
	if err != nil {
		return false, false, err
	}
	return result["password_changed"] == "1", result["route_changed"] == "1", nil
}

// AckTeamChangeNotice 清除用户明确确认过的团队变更通知类型。
func AckTeamChangeNotice(ctx context.Context, userID int64, passwordChanged, routeChanged bool) error {
	fields := make([]string, 0, 2)
	if passwordChanged {
		fields = append(fields, "password_changed")
	}
	if routeChanged {
		fields = append(fields, "route_changed")
	}
	if len(fields) == 0 {
		return nil
	}
	return client().HDel(ctx, buildTeamChangeNoticeKey(userID), fields...).Err()
}

func buildDailyTeamQuotaKey(day int) string {
	return dailyTeamQuotaKeyPrefix + strconv.Itoa(day)
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

func SubmitTeam(ctx context.Context, teamID int64, day int) (int64, error) {
	return submitTeamScript.Run(
		ctx,
		client(),
		[]string{
			submittedTeamsKey,
			buildDailyTeamQuotaKey(day),
			totalTeamQuotaKey,
			submittedTeamDaysKey,
		},
		teamID,
		day,
	).Int64()
}

func RollbackTeamSubmit(ctx context.Context, teamID int64, fallbackDay int) (bool, int, error) {
	teamIDValue := strconv.FormatInt(teamID, 10)
	result, err := rollbackTeamSubmitScript.Run(
		ctx,
		client(),
		[]string{submittedTeamsKey, submittedTeamDaysKey, totalTeamQuotaKey},
		teamIDValue,
		fallbackDay,
		dailyTeamQuotaKeyPrefix,
	).Int64Slice()
	if err != nil {
		return false, 0, err
	}
	if len(result) != 2 {
		return false, 0, fmt.Errorf("unexpected rollback team submit result: %v", result)
	}
	return result[0] == 1, int(result[1]), nil
}

func RestoreSubmittedTeam(ctx context.Context, teamID int64, day int) error {
	teamIDValue := strconv.FormatInt(teamID, 10)
	return restoreSubmittedTeamScript.Run(
		ctx,
		client(),
		[]string{
			submittedTeamsKey,
			submittedTeamDaysKey,
			buildDailyTeamQuotaKey(day),
			totalTeamQuotaKey,
		},
		teamIDValue,
		day,
	).Err()
}

func InitDailyTeamQuota(ctx context.Context, day int, limit int) error {
	key := buildDailyTeamQuotaKey(day)
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
