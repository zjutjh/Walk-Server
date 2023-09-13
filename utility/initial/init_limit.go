package initial

import (
	"github.com/juju/ratelimit"
	"time"
	"walk-server/global"
)

func LimitInit() {
	capacity := global.Config.GetInt64("QPS")
	quantum := capacity

	// func NewBucketWithQuantum(fillInterval time.Duration, capacity, quantum int64) *Bucket
	// fillInterval 指每过多长时间向桶里放一个令牌，capacity 是桶的容量，quantum 是每次向桶中放令牌的数量
	// 下面的意思是指任意一秒内最多可以接受的并发量为 capacity（即桶的容量）
	global.Bucket = ratelimit.NewBucketWithQuantum(time.Second, capacity, quantum)
}
