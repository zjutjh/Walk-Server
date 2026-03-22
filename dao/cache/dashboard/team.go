package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

const (
	teamInfoCacheKeyPrefix     = "dashboard:teams:info"
	teamInfoCacheTTL           = 60 * time.Second
	teamFilterCacheKeyPrefix   = "dashboard:teams:filter"
	teamFilterCacheTTL         = 30 * time.Second
	teamInfoLockCacheKeyPrefix = "dashboard:teams:info:lock"
)

// BuildTeamInfoCacheKey 构造队伍详情缓存 key。
func BuildTeamInfoCacheKey(teamID int64) string {
	return fmt.Sprintf("%s:%d", teamInfoCacheKeyPrefix, teamID)
}

// GetTeamInfo 返回缓存内容；命中返回 found=true，未命中返回 found=false。
func (c *DashboardCache) GetTeamInfo(ctx context.Context, teamID int64) (cached []byte, found bool, err error) {
	cached, err = c.rdb.Get(ctx, BuildTeamInfoCacheKey(teamID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return cached, true, nil
}

// SetTeamInfo 写入队伍详情缓存。
func (c *DashboardCache) SetTeamInfo(ctx context.Context, teamID int64, cached []byte) error {
	return c.rdb.Set(ctx, BuildTeamInfoCacheKey(teamID), cached, teamInfoCacheTTL).Err()
}

// DeleteTeamInfo 删除队伍详情缓存。
func (c *DashboardCache) DeleteTeamInfo(ctx context.Context, teamID int64) error {
	return c.rdb.Del(ctx, BuildTeamInfoCacheKey(teamID)).Err()
}

// BuildTeamFilterCacheKey 构造队伍筛选缓存 key。
func BuildTeamFilterCacheKey(campus string, queryHash string) string {
	return fmt.Sprintf("%s:%s:%s", teamFilterCacheKeyPrefix, campus, queryHash)
}

// GetTeamFilter 返回缓存内容；命中返回 found=true，未命中返回 found=false。
func (c *DashboardCache) GetTeamFilter(ctx context.Context, campus string, queryHash string) (cached []byte, found bool, err error) {
	cached, err = c.rdb.Get(ctx, BuildTeamFilterCacheKey(campus, queryHash)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return cached, true, nil
}

// SetTeamFilter 写入队伍筛选缓存。
func (c *DashboardCache) SetTeamFilter(ctx context.Context, campus string, queryHash string, cached []byte) error {
	return c.rdb.Set(ctx, BuildTeamFilterCacheKey(campus, queryHash), cached, teamFilterCacheTTL).Err()
}

func BuildTeamInfoLockCacheKey(teamID int64) string {
	return fmt.Sprintf("%s:%d", teamInfoLockCacheKeyPrefix, teamID)
}

// AcquireTeamInfoLock 尝试对队伍信息加锁，成功返回 true。
func (c *DashboardCache) AcquireTeamInfoLock(ctx context.Context, teamID int64, ttl time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, BuildTeamInfoLockCacheKey(teamID), 1, ttl).Result()
}

// SetTeamInfoLockTTL 覆盖写入队伍信息锁及过期时间。
func (c *DashboardCache) SetTeamInfoLockTTL(ctx context.Context, teamID int64, ttl time.Duration) error {
	if ttl <= 0 {
		return c.ReleaseTeamInfoLock(ctx, teamID)
	}

	return c.rdb.Set(ctx, BuildTeamInfoLockCacheKey(teamID), 1, ttl).Err()
}

// ReleaseTeamInfoLock 释放队伍信息锁。
func (c *DashboardCache) ReleaseTeamInfoLock(ctx context.Context, teamID int64) error {
	return c.rdb.Del(ctx, BuildTeamInfoLockCacheKey(teamID)).Err()
}
