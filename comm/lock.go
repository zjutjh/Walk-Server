package comm

import (
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/zjutjh/mygo/lock"
)

const (
	teamLockKeyPrefix = "walk:lock:team"
	teamLockExpiry    = 10 * time.Second
)

func NewTeamMutex(teamID int64) *redsync.Mutex {
	return lock.Pick().NewMutex(
		fmt.Sprintf("%s:%d", teamLockKeyPrefix, teamID),
		redsync.WithExpiry(teamLockExpiry),
		redsync.WithTries(1),
	)
}
