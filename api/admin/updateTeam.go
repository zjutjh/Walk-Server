package api

import (
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/session"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/model"
	"app/dao/query"
	repo "app/dao/repo"
)

func UpdateTeamHandler() gin.HandlerFunc {
	api := UpdateTeamApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(updateTeam).Pointer()).Name()] = api
	return updateTeam
}

type UpdateTeamApi struct {
	Info     struct{} `name:"打卡"`
	Request  UpdateTeamApiRequest
	Response UpdateTeamApiResponse
}

type UpdateTeamApiRequest struct {
	Body struct {
		CodeType string `json:"code_type" binding:"required"`
		Content  string `json:"content" binding:"required"`
	}
}

type UpdateTeamApiResponse struct {
	TeamID int `json:"team_id" desc:"队伍编号"`
}

type routePointCheckinResult struct {
	code *kit.Code
}

// Run Api业务逻辑执行点
func (u *UpdateTeamApi) Run(ctx *gin.Context) kit.Code {
	teamRepo := repo.NewTeamRepo()

	admin, code := u.getCurrentAdmin(ctx)
	if code != nil {
		return *code
	}

	team, code := u.resolveTeam(ctx, admin)
	if code != nil {
		return *code
	}

	mutex := comm.NewTeamMutex(team.ID)
	if err := mutex.Lock(); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取队伍打卡锁失败")
		return comm.CodeTooFrequently
	}
	defer func() {
		if _, err := mutex.Unlock(); err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("释放队伍打卡锁失败")
		}
	}()

	if err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		return repo.NewTeamRepoWithTx(tx).ClearLostStatus(ctx, team.ID)
	}); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("清除队伍失联状态失败")
		return comm.CodeDatabaseError
	}

	if team.PrevPointName == admin.PointName {
		return comm.CodeDuplicateCheckin
	}

	pointRoutes, err := teamRepo.FindPointRoutes(ctx, admin.PointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询点位所属路线失败")
		return comm.CodeDatabaseError
	}
	if len(pointRoutes) == 0 {
		return comm.CodeDataNotFound
	}

	if !slices.Contains(pointRoutes, team.RouteName) {
		result, err := u.handleWrongRoutePointCheckin(ctx, team, admin.ID, admin.PointName, pointRoutes[0])
		if err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Error("错路点位打卡失败")
			return comm.CodeDatabaseError
		}
		if result.code != nil {
			return *result.code
		}
		u.Response.TeamID = int(team.ID)
		return comm.CodeOK
	}

	routeEdge, err := teamRepo.FindRouteEdge(ctx, team.RouteName, admin.PointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线边失败")
		return comm.CodeDatabaseError
	}

	if routeEdge != nil && routeEdge.PrevPointName == "" {
		if team.Status != comm.TeamStatusNotStart {
			return comm.CodeTeamCheckinClosed
		}
		if err := u.handleStartPointCheckin(ctx, team, admin.ID, admin.PointName); err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Error("起点打卡失败")
			return comm.CodeDatabaseError
		}
		u.Response.TeamID = int(team.ID)
		return comm.CodeOK
	}

	result, err := u.handleRoutePointCheckin(ctx, team, admin.ID, admin.PointName, routeEdge)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("普通点位打卡失败")
		return comm.CodeDatabaseError
	}
	if result.code != nil {
		return *result.code
	}

	u.Response.TeamID = int(team.ID)
	return comm.CodeOK
}

func (u *UpdateTeamApi) getCurrentAdmin(ctx *gin.Context) (*model.Admin, *kit.Code) {
	adminRepo := repo.NewAdminRepo()

	adminID, err := session.GetIdentity[int64](ctx)
	if err != nil {
		adminIDInt, fallbackErr := session.GetIdentity[int](ctx)
		if fallbackErr != nil {
			nlog.Pick().WithContext(ctx).WithError(fallbackErr).Warn("获取管理员登录态失败")
			return nil, &comm.CodeNotLoggedIn
		}
		adminID = int64(adminIDInt)
	}

	admin, err := adminRepo.FindByID(ctx, adminID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询管理员失败")
		return nil, &comm.CodeDatabaseError
	}
	if admin == nil {
		return nil, &comm.CodeNotLoggedIn
	}
	return admin, nil
}

func (u *UpdateTeamApi) resolveTeam(ctx *gin.Context, admin *model.Admin) (*model.Team, *kit.Code) {
	teamRepo := repo.NewTeamRepo()

	content := strings.TrimSpace(u.Request.Body.Content)
	codeType := strings.TrimSpace(u.Request.Body.CodeType)

	var (
		team *model.Team
		err  error
	)

	switch codeType {
	case comm.CodeTeam:
		teamID, parseErr := strconv.ParseInt(content, 10, 64)
		if parseErr != nil {
			return nil, &comm.CodeParameterInvalid
		}
		team, err = teamRepo.FindTeamByID(ctx, teamID)
	case comm.CodeCheckin:
		team, err = teamRepo.FindByCode(ctx, content)
	default:
		return nil, &comm.CodeParameterInvalid
	}
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍失败")
		return nil, &comm.CodeDatabaseError
	}
	if team == nil {
		return nil, &comm.CodeTeamNotFound
	}

	route, err := teamRepo.FindRouteByName(ctx, team.RouteName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线失败")
		return nil, &comm.CodeDatabaseError
	}
	if route == nil {
		return nil, &comm.CodeDataNotFound
	}
	if route.Campus != admin.Campus {
		return nil, &comm.CodeCampusMismatch
	}
	return team, nil
}

func (u *UpdateTeamApi) handlePointCheckin(ctx *gin.Context, team *model.Team, adminID int64, pointName string) error {
	return query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		if err := txTeamRepo.UpdatePrevPointName(ctx, team.ID, pointName); err != nil {
			return err
		}
		return txTeamRepo.CreateCheckin(ctx, adminID, team.ID, pointName, team.RouteName)
	})
}

func (u *UpdateTeamApi) handleStartPointCheckin(ctx *gin.Context, team *model.Team, adminID int64, pointName string) error {
	return query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		teamRepo := repo.NewTeamRepoWithTx(tx)
		peopleRepo := repo.NewPeopleRepoWithTx(tx)
		if err := teamRepo.UpdatePrevPointName(ctx, team.ID, pointName); err != nil {
			return err
		}
		if err := teamRepo.CreateCheckin(ctx, adminID, team.ID, pointName, team.RouteName); err != nil {
			return err
		}
		return peopleRepo.UpdateByTeamID(ctx, team.ID, map[string]any{"walk_status": comm.WalkStatusPending})
	})
}

func (u *UpdateTeamApi) handleRoutePointCheckin(ctx *gin.Context, team *model.Team, adminID int64, pointName string, routeEdge *model.RouteEdge) (*routePointCheckinResult, error) {
	if team.Status != comm.TeamStatusInProgress {
		return &routePointCheckinResult{code: &comm.CodeTeamCheckinClosed}, nil
	}

	if routeEdge == nil || routeEdge.PrevPointName != team.PrevPointName {
		return &routePointCheckinResult{code: &comm.CodePrevPointInvalid}, nil
	}

	if err := u.handlePointCheckin(ctx, team, adminID, pointName); err != nil {
		return nil, err
	}

	return &routePointCheckinResult{}, nil
}

func (u *UpdateTeamApi) handleWrongRoutePointCheckin(ctx *gin.Context, team *model.Team, adminID int64, pointName string, wrongRouteName string) (*routePointCheckinResult, error) {
	if team.Status != comm.TeamStatusInProgress {
		return &routePointCheckinResult{code: &comm.CodeTeamCheckinClosed}, nil
	}

	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		if err := txTeamRepo.UpdatePrevPointName(ctx, team.ID, pointName); err != nil {
			return err
		}
		if err := txTeamRepo.CreateCheckin(ctx, adminID, team.ID, pointName, team.RouteName); err != nil {
			return err
		}
		if err := txTeamRepo.UpdateTeamWrongRoute(ctx, team.ID, 1); err != nil {
			return err
		}
		return txTeamRepo.CreateWrongRouteRecord(ctx, team.ID, team.RouteName, wrongRouteName, adminID)
	})
	if err != nil {
		return nil, err
	}

	return &routePointCheckinResult{code: &comm.CodeWrongRouteAlert}, nil
}

// Run Api初始化 进行参数校验和绑定
func (u *UpdateTeamApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&u.Request.Body)
	if err != nil {
		return err
	}
	return err
}

// updateTeam Api执行入口
func updateTeam(ctx *gin.Context) {
	api := &UpdateTeamApi{}
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
