package api

import (
	"errors"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	peopleCache "app/dao/cache/people"
	teamCache "app/dao/cache/team"
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

var (
	errBindTeamNotEnough = errors.New("bind code team not enough")
	errBindTeamFull      = errors.New("bind code team full")
	errBindCodeDuplicate = errors.New("bind code duplicate")
	errBindCodeEmpty     = errors.New("bind code empty")
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

	newCode := strings.TrimSpace(b.Request.Body.Content)
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)
		txAdminRepo := repo.NewAdminRepoWithTx(tx)
		if newCode == "" {
			return errBindCodeEmpty
		}

		if team.Code == newCode {
			return nil
		}

		codeOwner, err := txTeamRepo.FindByCode(ctx, newCode)
		if err != nil {
			return err
		}
		if codeOwner != nil && codeOwner.ID != team.ID {
			return errBindCodeDuplicate
		}

		startEdge, err := txTeamRepo.FindRouteStartEdge(ctx, team.RouteName)
		if err != nil {
			return err
		}
		if startEdge == nil {
			return errors.New("route start point not found")
		}

		checkedIn, err := txTeamRepo.HasTeamCheckinAtPoint(ctx, team.ID, startEdge.PointName)
		if err != nil {
			return err
		}
		if !checkedIn {
			admin, err := txAdminRepo.FindByPointName(ctx, startEdge.PointName)
			if err != nil {
				return err
			}
			if admin == nil {
				return errors.New("start point admin not found")
			}

			if err := txTeamRepo.UpdateLatestPointName(ctx, team.ID, startEdge.PointName); err != nil {
				return err
			}
			if err := txTeamRepo.CreateCheckin(ctx, admin.ID, team.ID, startEdge.PointName, team.RouteName); err != nil {
				return err
			}
			if err := txPeopleRepo.UpdateMembersWalkStatusByCurrent(ctx, team.ID, comm.WalkStatusNotStart, comm.WalkStatusPending); err != nil {
				return err
			}
		}

		pendingCount, err := txPeopleRepo.CountMembersByStatus(ctx, team.ID, comm.WalkStatusPending)
		if err != nil {
			return err
		}
		if pendingCount < minTeamMemberCount {
			return errBindTeamNotEnough
		}
		if pendingCount > maxTeamMemberCount {
			return errBindTeamFull
		}

		return txTeamRepo.UpdateCodeByID(ctx, team.ID, newCode)
	})
	if err != nil {
		if errors.Is(err, errBindCodeEmpty) {
			return comm.CodeParameterInvalid
		}
		if errors.Is(err, errBindTeamNotEnough) {
			return comm.CodeTeamNotEnough
		}
		if errors.Is(err, errBindTeamFull) {
			return comm.CodeTeamFull
		}
		if errors.Is(err, errBindCodeDuplicate) {
			return comm.CodeBindCodeDuplicated
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("绑定签到码失败")
		return comm.CodeBindCodeError
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	if team.Code != "" {
		_ = teamCache.DelTeamIDByCode(ctx, team.Code)
	}
	if newCode != "" {
		_ = teamCache.DelTeamIDByCode(ctx, newCode)
	}
	if members, err := repo.NewPeopleRepo().FindPeopleByTeamID(ctx, team.ID); err == nil {
		for _, member := range members {
			if member != nil && member.OpenID != "" {
				_ = peopleCache.DelPersonByOpenID(ctx, member.OpenID)
			}
		}
	}

	return comm.CodeOK
}

func (b *BindCodeApi) getTeam(ctx *gin.Context) (*model.Team, *kit.Code) {
	teamRepo := repo.NewTeamRepo()

	team, err := teamRepo.FindTeamByID(ctx, int64(b.Request.Body.TeamID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍失败")
		return nil, &comm.CodeServerError
	}
	if team == nil {
		return nil, &comm.CodeTeamNotFound
	}
	return team, nil
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
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
