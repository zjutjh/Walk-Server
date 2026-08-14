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
	submittedTeamDaysKey       = "walk:team:submitted_days"
	totalTeamQuotaKey          = "walk:team:quota:total"
	teamInfoCacheKeyPrefix     = "dashboard:teams:info"
	teamInfoCacheTTL           = 60 * time.Second
	teamFilterCacheKeyPrefix   = "dashboard:teams:filter"
	teamFilterCacheTTL         = 30 * time.Second
	teamInfoLockCacheKeyPrefix = "dashboard:teams:info:lock"
)

var teamInfoLocks sync.Map

var submitTeamScript = redis.NewScript(`
local teamID = KEYS[1]
local dailyQuotaKey = KEYS[2]
local totalQuotaKey = KEYS[3]
local submittedDaysKey = KEYS[4]
local day = ARGV[1]

local submitted = redis.call("SISMEMBER", "teams", teamID)
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

redis.call("SADD", "teams", teamID)
redis.call("HSET", submittedDaysKey, teamID, day)
redis.call("DECR", dailyQuotaKey)
redis.call("DECR", totalQuotaKey)
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

func buildDailyTeamQuotaKey(day int) string {
	return "walk:team:quota:day:" + strconv.Itoa(day)
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

func SubmitTeam(ctx context.Context, teamID int64, day int) (int64, error) {
	return submitTeamScript.Run(
		ctx,
		client(),
		[]string{
			strconv.FormatInt(teamID, 10),
			buildDailyTeamQuotaKey(day),
			totalTeamQuotaKey,
			submittedTeamDaysKey,
		},
		day,
	).Int64()
}

func RollbackTeamSubmit(ctx context.Context, teamID int64, fallbackDay int) (bool, int, error) {
	teamIDValue := strconv.FormatInt(teamID, 10)
	submitted, err := client().SIsMember(ctx, submittedTeamsKey, teamIDValue).Result()
	if err != nil {
		return false, 0, err
	}
	if !submitted {
		return false, 0, nil
	}
	day := fallbackDay
	if value, err := client().HGet(ctx, submittedTeamDaysKey, teamIDValue).Int(); err == nil {
		day = value
	} else if err != redis.Nil {
		return false, 0, err
	}
	if err := client().SRem(ctx, submittedTeamsKey, teamIDValue).Err(); err != nil {
		return false, 0, err
	}
	_ = client().HDel(ctx, submittedTeamDaysKey, teamIDValue).Err()
	if err := client().Incr(ctx, buildDailyTeamQuotaKey(day)).Err(); err != nil {
		return false, 0, err
	}
	if err := client().Incr(ctx, totalTeamQuotaKey).Err(); err != nil {
		return false, 0, err
	}
	return true, day, nil
}

func RestoreSubmittedTeam(ctx context.Context, teamID int64, day int) error {
	teamIDValue := strconv.FormatInt(teamID, 10)
	if err := client().SAdd(ctx, submittedTeamsKey, teamIDValue).Err(); err != nil {
		return err
	}
	if err := client().HSet(ctx, submittedTeamDaysKey, teamIDValue, day).Err(); err != nil {
		return err
	}
	if err := client().Decr(ctx, buildDailyTeamQuotaKey(day)).Err(); err != nil {
		return err
	}
	return client().Decr(ctx, totalTeamQuotaKey).Err()
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
