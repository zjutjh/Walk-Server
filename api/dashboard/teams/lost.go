package teams

import (
	"errors"
	"reflect"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"gorm.io/gorm"

	"app/comm"
	teamCache "app/dao/cache/team"
	repo "app/dao/repo"
)

const lostUpdateLockTTL = 5 * time.Minute

// LostHandler API router注册点
func LostHandler() gin.HandlerFunc {
	api := LostApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfLost).Pointer()).Name()] = api
	return hfLost
}

type LostApi struct {
	Info     struct{}        `name:"设置队伍失联状态" desc:"设置指定队伍的失联状态 \n 距现在5min内打卡的队伍不允许设置失联状态为true \n 设置为false不受时间限制，但会覆盖之前的失联状态和时间"`
	Request  LostApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response LostApiResponse // API响应数据 (Body中的Data部分)
}

type LostApiRequest struct {
	Body struct {
		TeamId int64 `json:"team_id" desc:"队伍ID"`
		IsLost bool  `json:"is_lost" desc:"是否失联"`
	}
}

type LostApiResponse struct{}

// Run Api业务逻辑执行点
func (l *LostApi) Run(ctx *gin.Context) kit.Code {
	teamID := l.Request.Body.TeamId
	if teamID <= 0 {
		return comm.CodeParameterInvalid
	}

	var lockAcquired bool
	var keepLock bool
	// 仅当 is_lost=true 时需要加锁
	if l.Request.Body.IsLost {
		var lockErr error
		lockAcquired, lockErr = teamCache.AcquireTeamInfoLock(ctx, teamID, lostUpdateLockTTL)
		if lockErr != nil {
			nlog.Pick().WithContext(ctx).WithError(lockErr).Warn("队伍失联状态加锁失败，降级走数据库校验")
		}
		if lockAcquired == false && lockErr == nil {
			return comm.CodeTooFrequently
		}
	}

	keepLock = lockAcquired
	defer func() {
		if !lockAcquired || keepLock {
			return
		}
		releaseErr := teamCache.ReleaseTeamInfoLock(ctx, teamID)
		if releaseErr != nil {
			nlog.Pick().WithContext(ctx).WithError(releaseErr).Warn("释放队伍失联状态锁失败")
		}
	}()

	teamRepo := repo.NewTeamRepo()
	team, err := teamRepo.GetTeamByID(ctx, teamID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		keepLock = false
		return comm.CodeDataNotFound
	}
	if err != nil {
		keepLock = false
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍信息失败")
		return comm.CodeDatabaseError
	}

	// 5min锁定：仅当 is_lost=true 时，检查队伍状态更新时间在5分钟内是否不允许重复更新
	now := time.Now()
	if l.Request.Body.IsLost && !team.Time.IsZero() && now.Before(team.Time.Add(lostUpdateLockTTL)) {
		remaining := time.Until(team.Time.Add(lostUpdateLockTTL))
		if remaining > 0 {
			setErr := teamCache.SetTeamInfoLockTTL(ctx, teamID, remaining)
			if setErr != nil {
				nlog.Pick().WithContext(ctx).WithError(setErr).Warn("回写队伍失联状态锁失败")
			}
		}
		keepLock = true
		return comm.CodeTooFrequently
	}

	updated, err := teamRepo.UpdateTeamLostStatus(ctx, teamID, l.Request.Body.IsLost, now)
	if err != nil {
		keepLock = false
		nlog.Pick().WithContext(ctx).WithError(err).Error("更新队伍失联状态失败")
		return comm.CodeDatabaseError
	}
	if !updated {
		keepLock = false
		return comm.CodeDataNotFound
	}

	if lockAcquired {
		keepLock = true
	}

	err = teamCache.DeleteTeamInfo(ctx, teamID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("删除队伍详情缓存失败")
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (l *LostApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&l.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// hfLost API执行入口
func hfLost(ctx *gin.Context) {
	api := &LostApi{}
	err := api.Init(ctx)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Success(ctx, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
