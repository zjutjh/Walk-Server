package cache

import (
	"context"
	"fmt"
	"time"
)

const teamInfoLockCacheKeyPrefix = "dashboard:teams:info:lock"

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
