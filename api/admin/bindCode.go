package api

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/model"
	"app/dao/query"
	repo "app/dao/repo"
)

func BindCodeHandler() gin.HandlerFunc {
	api := BindCodeApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(bindCode).Pointer()).Name()] = api
	return bindCode
}

type BindCodeApi struct {
	Info     struct{} `name:"绑定签到码"`
	Request  BindCodeApiRequest
	Response BindCodeApiResponse
}

type BindCodeApiRequest struct {
	Body struct {
		TeamID  int    `json:"team_id" desc:"团队编号" binding:"required"`
		Content string `json:"content" desc:"签到码" binding:"required"`
	}
}

type BindCodeApiResponse struct {
}

const (
	minTeamMemberCount = 3
	maxTeamMemberCount = 6
)

// Run Api业务逻辑执行点
func (b *BindCodeApi) Run(ctx *gin.Context) kit.Code {
	team, code := b.getTeam(ctx)
	if code != nil {
		return *code
	}

	mutex := comm.NewTeamMutex(team.ID)
	if err := mutex.Lock(); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取队伍绑定签到码锁失败")
		return comm.CodeTooFrequently
	}
	defer func() {
		if _, err := mutex.Unlock(); err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("释放队伍绑定签到码锁失败")
		}
	}()

	code = b.validatePendingMemberCount(ctx, team.ID)
	if code != nil {
		return *code
	}

	err := b.bindCode(ctx, team.ID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("绑定签到码失败")
		return comm.CodeBindCodeError
	}

	return comm.CodeOK
}

func (b *BindCodeApi) getTeam(ctx *gin.Context) (*model.Team, *kit.Code) {
	teamRepo := repo.NewTeamRepo()

	team, err := teamRepo.FindTeamByID(ctx, int64(b.Request.Body.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍失败")
		return nil, &comm.CodeDatabaseError
	}
	if team == nil {
		return nil, &comm.CodeTeamNotFound
	}
	return team, nil
}

func (b *BindCodeApi) validatePendingMemberCount(ctx *gin.Context, teamID int64) *kit.Code {
	peopleRepo := repo.NewPeopleRepo()

	pendingCount, err := peopleRepo.CountMembersByStatus(ctx, teamID, comm.WalkStatusPending)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("统计待出发人数失败")
		return &comm.CodeDatabaseError
	}
	if pendingCount < minTeamMemberCount {
		return &comm.CodeTeamNotEnough
	}
	if pendingCount > maxTeamMemberCount {
		return &comm.CodeTeamFull
	}
	return nil
}

func (b *BindCodeApi) bindCode(ctx *gin.Context, teamID int64) error {
	return query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)

		if err := txTeamRepo.UpdateByID(ctx, teamID, map[string]any{"code": b.Request.Body.Content}); err != nil {
			return err
		}
		if err := txPeopleRepo.UpdateMembersWalkStatusByCurrent(ctx, teamID, comm.WalkStatusPending, comm.WalkStatusInProgress); err != nil {
			return err
		}

		inProgressCount, err := txPeopleRepo.CountMembersByStatus(ctx, teamID, comm.WalkStatusInProgress)
		if err != nil {
			return err
		}
		if inProgressCount > 0 {
			if err := txTeamRepo.UpdateByID(ctx, teamID, map[string]any{"status": comm.TeamStatusInProgress}); err != nil {
				return err
			}
		}
		return nil
	})
}

// Run Api初始化 进行参数校验和绑定
func (b *BindCodeApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&b.Request.Body)
	if err != nil {
		return err
	}
	return err
}

func bindCode(ctx *gin.Context) {
	api := &BindCodeApi{}
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
