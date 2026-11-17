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
	teamCache "app/dao/cache/team"
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
		CodeType string `json:"code_type" desc:"码类型：checkin签到码，team团队码，user个人码" binding:"required"`
		Content  string `json:"content" desc:"码内容；团队码为团队ID，个人码为用户ID" binding:"required"`
	}
}

type UpdateTeamApiResponse struct {
	TeamID    int    `json:"team_id" desc:"队伍编号"`
	Exception string `json:"exception" desc:"非阻断性异常类型，空字符串表示无异常，duplicate表示重复打卡，wrong_direction表示行进方向异常"`
}

const (

	checkinEventNormal             = "normal"
	checkinEventSkippedPoint       = "skipped_point"
	checkinEventWrongDirection     = "wrong_direction"
	checkinEventDuplicate          = "duplicate"
	checkinExceptionNone           = ""
	checkinExceptionDuplicate      = "duplicate"
	checkinExceptionWrongDirection = "wrong_direction"
)

type checkinOptions struct {
	advance        bool
	startPoint     bool
	wrongRouteName string
	recovering     bool
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

	if err := teamRepo.ClearLostStatus(ctx, team.ID); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("清除队伍失联状态失败")
		return comm.CodeServerError
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)

	pointRoutes, err := teamRepo.FindPointRoutes(ctx, admin.PointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询点位所属路线失败")
		return comm.CodeServerError
	}
	if len(pointRoutes) == 0 {
		return comm.CodeDataNotFound
	}

	activeRouteName, directionInvalid, directionCode := u.resolveActiveRouteForDirection(ctx, teamRepo, team, pointRoutes)
	if directionCode != nil {
		return *directionCode
	}

	recovering := team.IsWrongRoute && slices.Contains(pointRoutes, team.RouteName)
	enteringWrongRoute := !slices.Contains(pointRoutes, team.RouteName) && !team.IsWrongRoute

	options := checkinOptions{advance: true, recovering: recovering}
	if enteringWrongRoute {
		options.wrongRouteName = activeRouteName
	}

	if directionInvalid {
		if team.IsWrongRoute && !recovering {
			options.wrongRouteName = activeRouteName
		}
		return u.recordCheckinException(ctx, team, admin, activeRouteName, checkinEventWrongDirection, checkinExceptionWrongDirection, options)
	}

	routeEdge, err := teamRepo.FindRouteTransitionEdge(ctx, activeRouteName, team.LatestPointName, admin.PointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线边失败")
		return comm.CodeServerError
	}
	sequence := repo.RouteCheckinSequence{}
	if routeEdge == nil && team.LatestPointName != "" {
		sequence, err = teamRepo.FindRouteCheckinSequence(ctx, activeRouteName, team.LatestPointName, admin.PointName)
		if err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Error("查询打卡点位路线序号失败")
			return comm.CodeServerError
		}
	}
	eventType := classifyCheckinEvent(team.LatestPointName, admin.PointName, sequence, routeEdge != nil)

	if eventType == checkinEventWrongDirection || eventType == checkinEventDuplicate {
		exception := checkinExceptionWrongDirection
		if eventType == checkinEventDuplicate {
			exception = checkinExceptionDuplicate
		}
		return u.recordCheckinException(ctx, team, admin, activeRouteName, eventType, exception, options)
	}

	options.startPoint = options.startPoint || routeEdge != nil && routeEdge.PrevPointName == ""

	if err := u.handleCheckin(ctx, team, admin.ID, admin.PointName, activeRouteName, eventType, options); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("普通点位打卡失败")
		return comm.CodeServerError
	}

	u.Response.TeamID = int(team.ID)
	return comm.CodeOK
}

func (u *UpdateTeamApi) recordCheckinException(ctx *gin.Context, team *model.Team, admin *model.Admin, routeName, eventType, exception string, options checkinOptions) kit.Code {
	if err := u.handleCheckin(ctx, team, admin.ID, admin.PointName, routeName, eventType, options); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("记录异常打卡失败")
		return comm.CodeServerError
	}
	u.Response.TeamID = int(team.ID)
	u.Response.Exception = exception
	return comm.CodeOK
}

func classifyCheckinEvent(previousPoint, currentPoint string, sequence repo.RouteCheckinSequence, hasDirectEdge bool) string {
	if previousPoint != "" && previousPoint == currentPoint {
		return checkinEventDuplicate
	}
	if hasDirectEdge {
		return checkinEventNormal
	}
	if previousPoint == "" {
		return checkinEventSkippedPoint
	}
	if !sequence.HasPrevious || !sequence.HasCurrent {
		return checkinEventNormal
	}
	if sequence.CurrentSeq == sequence.PreviousSeq {
		return checkinEventDuplicate
	}
	if sequence.CurrentSeq < sequence.PreviousSeq {
		return checkinEventWrongDirection
	}
	if sequence.HasNext {
		if sequence.CurrentSeq == sequence.NextSeq {
			return checkinEventNormal
		}
		if sequence.CurrentSeq > sequence.NextSeq {
			return checkinEventSkippedPoint
		}
		return checkinEventWrongDirection
	}
	if sequence.CurrentSeq > sequence.PreviousSeq+1 {
		return checkinEventSkippedPoint
	}
	return checkinEventNormal
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
		return nil, &comm.CodeServerError
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
	case comm.CodeUser:
		userID, parseErr := strconv.ParseInt(content, 10, 64)
		if parseErr != nil || userID <= 0 {
			return nil, &comm.CodeParameterInvalid
		}
		person, findErr := repo.NewPeopleRepo().FindPeopleByID(ctx, userID)
		if findErr != nil {
			nlog.Pick().WithContext(ctx).WithError(findErr).Error("通过个人码查询人员失败")
			return nil, &comm.CodeServerError
		}
		if person == nil {
			return nil, &comm.CodePeopleNotFound
		}
		if person.TeamID <= 0 {
			return nil, &comm.CodeNotInTeam
		}
		team, err = teamRepo.FindTeamByID(ctx, person.TeamID)
	case comm.CodeCheckin:
		team, err = teamRepo.FindByCode(ctx, content)
	default:
		return nil, &comm.CodeParameterInvalid
	}
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍失败")
		return nil, &comm.CodeServerError
	}
	if team == nil {
		return nil, &comm.CodeTeamNotFound
	}

	route, err := teamRepo.FindRouteByName(ctx, team.RouteName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路线失败")
		return nil, &comm.CodeServerError
	}
	if route == nil {
		return nil, &comm.CodeDataNotFound
	}
	if route.Campus != admin.Campus {
		return nil, &comm.CodeCampusMismatch
	}
	return team, nil
}

func (u *UpdateTeamApi) resolveActiveRouteForDirection(ctx *gin.Context, teamRepo *repo.TeamRepo, team *model.Team, pointRoutes []string) (string, bool, *kit.Code) {
	activeRoute := pointRoutes[0]
	if slices.Contains(pointRoutes, team.RouteName) {
		activeRoute = team.RouteName
	}
	if team.LatestPointName == "" {
		return activeRoute, false, nil
	}

	previousRoutes, err := teamRepo.FindPointRoutes(ctx, team.LatestPointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询前一点位所属路线失败")
		return "", false, &comm.CodeServerError
	}
	sharedRoute := ""
	for _, routeName := range pointRoutes {
		if slices.Contains(previousRoutes, routeName) {
			sharedRoute = routeName
			break
		}
	}
	if sharedRoute != "" && !slices.Contains(pointRoutes, team.RouteName) {
		activeRoute = sharedRoute
	}
	if team.IsWrongRoute {
		wrongRouteName, found, err := teamRepo.GetLatestWrongRouteName(ctx, team.ID)
		if err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍错路路线失败")
			return "", false, &comm.CodeServerError
		}
		if found && slices.Contains(pointRoutes, wrongRouteName) {
			activeRoute = wrongRouteName
		}
	}
	return activeRoute, sharedRoute == "", nil
}

func (u *UpdateTeamApi) handleCheckin(ctx *gin.Context, team *model.Team, adminID int64, pointName, routeName, eventType string, options checkinOptions) error {
	err := query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		peopleRepo := repo.NewPeopleRepoWithTx(tx)
		if options.advance {
			if err := txTeamRepo.UpdateLatestPointName(ctx, team.ID, pointName); err != nil {
				return err
			}
		}
		if err := txTeamRepo.CreateCheckin(ctx, adminID, team.ID, team.LatestPointName, pointName, routeName, eventType); err != nil {
			return err
		}
		if options.wrongRouteName != "" {
			if err := txTeamRepo.UpdateTeamWrongRoute(ctx, team.ID, true); err != nil {
				return err
			}
			if err := txTeamRepo.CreateWrongRouteRecord(ctx, team.ID, team.RouteName, options.wrongRouteName, adminID); err != nil {
				return err
			}
		}
		if options.recovering {
			if err := txTeamRepo.UpdateTeamWrongRoute(ctx, team.ID, false); err != nil {
				return err
			}
		}
		if options.startPoint {
			return peopleRepo.UpdateMembersWalkStatusByCurrent(ctx, team.ID, comm.WalkStatusNotStart, comm.WalkStatusPending)
		}
		if !options.advance || eventType == checkinEventDuplicate {
			return nil
		}
		return peopleRepo.UpdateMembersWalkStatusByCurrent(ctx, team.ID, comm.WalkStatusPending, comm.WalkStatusInProgress)
	})
	if err != nil {
		return err
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	return nil
}

// Run Api初始化 进行参数校验和绑定
func (u *UpdateTeamApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&u.Request.Body)
}

// updateTeam Api执行入口
func updateTeam(ctx *gin.Context) {
	api := &UpdateTeamApi{}
	if err := api.Init(ctx); err != nil {
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
