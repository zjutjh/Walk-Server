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
	"gorm.io/gorm"

	"app/comm"
	"app/dao/repo"
)

func TeamCreateHandler() gin.HandlerFunc {
	api := TeamCreateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamCreate).Pointer()).Name()] = api
	return hfTeamCreate
}

func TeamJoinHandler() gin.HandlerFunc {
	api := TeamJoinApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamJoin).Pointer()).Name()] = api
	return hfTeamJoin
}

func TeamInfoHandler() gin.HandlerFunc {
	api := TeamInfoApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamInfo).Pointer()).Name()] = api
	return hfTeamInfo
}

func TeamUpdateHandler() gin.HandlerFunc {
	api := TeamUpdateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamUpdate).Pointer()).Name()] = api
	return hfTeamUpdate
}

func TeamLeaveHandler() gin.HandlerFunc {
	api := TeamLeaveApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamLeave).Pointer()).Name()] = api
	return hfTeamLeave
}

func TeamDisbandHandler() gin.HandlerFunc {
	api := TeamDisbandApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamDisband).Pointer()).Name()] = api
	return hfTeamDisband
}

type TeamCreateApi struct {
	Info     struct{} `name:"创建队伍" desc:"创建队伍"`
	Request  TeamCreateApiRequest
	Response TeamCreateApiResponse
}

type TeamCreateApiRequest struct {
	Name       string `json:"name" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Slogan     string `json:"slogan"`
	AllowMatch bool   `json:"allow_match"`
	RouteID    int64  `json:"route_id"`
}

type TeamCreateApiResponse struct {
	TeamID int64 `json:"team_id"`
}

type TeamJoinApi struct {
	Info     struct{} `name:"加入队伍" desc:"加入已有队伍"`
	Request  TeamJoinApiRequest
	Response struct{}
}

type TeamJoinApiRequest struct {
	TeamID   int64  `json:"team_id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TeamInfoApi struct {
	Info     struct{} `name:"队伍信息" desc:"获取当前队伍与成员信息"`
	Request  struct{}
	Response TeamInfoApiResponse
}

type TeamInfoApiResponse struct {
	Team    *repo.Team    `json:"team"`
	Members []repo.Person `json:"members"`
}

type TeamUpdateApi struct {
	Info     struct{} `name:"修改队伍" desc:"队长修改队伍信息"`
	Request  TeamUpdateApiRequest
	Response struct{}
}

type TeamUpdateApiRequest struct {
	Name       string `json:"name" binding:"required"`
	Password   string `json:"password" binding:"required"`
	Slogan     string `json:"slogan"`
	AllowMatch bool   `json:"allow_match"`
	RouteID    int64  `json:"route_id"`
}

type TeamLeaveApi struct {
	Info     struct{} `name:"退出队伍" desc:"普通队员退出队伍"`
	Request  struct{}
	Response struct{}
}

type TeamDisbandApi struct {
	Info     struct{} `name:"解散队伍" desc:"队长解散队伍"`
	Request  struct{}
	Response struct{}
}

func (h *TeamCreateApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *TeamJoinApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *TeamInfoApi) Init(ctx *gin.Context) (err error) {
	return err
}

func (h *TeamUpdateApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *TeamLeaveApi) Init(ctx *gin.Context) (err error) {
	return err
}

func (h *TeamDisbandApi) Init(ctx *gin.Context) (err error) {
	return err
}

func (h *TeamCreateApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil {
		return comm.CodeDataNotFound
	}
	if person.TeamID > 0 {
		return comm.CodeAlreadyInTeam
	}
	if person.CreatedOp == 0 {
		return comm.CodeNoCreateChance
	}

	duplicated, err := teamRepo.FindByName(ctx, h.Request.Name)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if duplicated != nil {
		return comm.CodeTeamNameDuplicated
	}

	err = ndb.Pick().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txPeopleRepo := repo.NewPeopleRepoWithDB(tx)
		txTeamRepo := repo.NewTeamRepoWithDB(tx)

		team := &repo.Team{
			Name:       h.Request.Name,
			Num:        1,
			Password:   h.Request.Password,
			Slogan:     h.Request.Slogan,
			AllowMatch: h.Request.AllowMatch,
			Captain:    openID,
			RouteID:    h.Request.RouteID,
			PointID:    0,
			StartNum:   0,
			Submit:     false,
			Status:     1,
			Code:       "",
			IsLost:     false,
		}
		if errTx := txTeamRepo.Create(ctx, team); errTx != nil {
			return errTx
		}

		if errTx := txPeopleRepo.UpdateByOpenID(ctx, openID, map[string]any{
			"team_id":     team.ID,
			"role":        2,
			"created_op":  person.CreatedOp - 1,
			"walk_status": 1,
		}); errTx != nil {
			return errTx
		}

		h.Response.TeamID = team.ID
		return nil
	})
	if err != nil {
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func (h *TeamJoinApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil {
		return comm.CodeDataNotFound
	}
	if person.TeamID > 0 {
		return comm.CodeAlreadyInTeam
	}
	if person.JoinOp == 0 {
		return comm.CodeNoJoinChance
	}

	team, err := teamRepo.FindByID(ctx, h.Request.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}
	if team.Password != h.Request.Password {
		return comm.CodePasswordWrong
	}
	if team.Submit {
		return comm.CodeTeamSubmitted
	}
	maxTeamSize := 6
	if comm.BizConf != nil && comm.BizConf.MaxTeamSize > 0 {
		maxTeamSize = comm.BizConf.MaxTeamSize
	}
	if int(team.Num) >= maxTeamSize {
		return comm.CodeTeamFull
	}

	err = ndb.Pick().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txPeopleRepo := repo.NewPeopleRepoWithDB(tx)
		txTeamRepo := repo.NewTeamRepoWithDB(tx)

		if errTx := txTeamRepo.UpdateByID(ctx, team.ID, map[string]any{
			"num": int(team.Num) + 1,
		}); errTx != nil {
			return errTx
		}
		if errTx := txPeopleRepo.UpdateByOpenID(ctx, openID, map[string]any{
			"team_id": team.ID,
			"role":    1,
			"join_op": person.JoinOp - 1,
		}); errTx != nil {
			return errTx
		}
		return nil
	})
	if err != nil {
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func (h *TeamInfoApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}

	team, err := teamRepo.FindByID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}

	members, err := peopleRepo.ListByTeamID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}

	h.Response.Team = team
	h.Response.Members = members
	return comm.CodeOK
}

func (h *TeamUpdateApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}
	if person.Role != 2 {
		return comm.CodeNotCaptain
	}

	team, err := teamRepo.FindByID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}
	if team.Submit {
		return comm.CodeTeamSubmitted
	}

	err = teamRepo.UpdateByID(ctx, team.ID, map[string]any{
		"name":        h.Request.Name,
		"password":    h.Request.Password,
		"slogan":      h.Request.Slogan,
		"allow_match": h.Request.AllowMatch,
		"route_id":    h.Request.RouteID,
	})
	if err != nil {
		return comm.CodeDatabaseError
	}
	return comm.CodeOK
}

func (h *TeamLeaveApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}
	if person.Role == 2 {
		return comm.CodeNotCaptain
	}

	team, err := teamRepo.FindByID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}
	if team.Submit {
		return comm.CodeTeamSubmitted
	}

	err = ndb.Pick().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txPeopleRepo := repo.NewPeopleRepoWithDB(tx)
		txTeamRepo := repo.NewTeamRepoWithDB(tx)

		if errTx := txTeamRepo.UpdateByID(ctx, team.ID, map[string]any{
			"num": int(team.Num) - 1,
		}); errTx != nil {
			return errTx
		}
		if errTx := txPeopleRepo.UpdateByOpenID(ctx, openID, map[string]any{
			"team_id": -1,
			"role":    0,
		}); errTx != nil {
			return errTx
		}
		return nil
	})
	if err != nil {
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func (h *TeamDisbandApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if person == nil || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}
	if person.Role != 2 {
		return comm.CodeNotCaptain
	}

	team, err := teamRepo.FindByID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}
	if team.Submit {
		return comm.CodeTeamSubmitted
	}

	err = ndb.Pick().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txPeopleRepo := repo.NewPeopleRepoWithDB(tx)
		txTeamRepo := repo.NewTeamRepoWithDB(tx)

		if errTx := txPeopleRepo.UpdateByTeamID(ctx, team.ID, map[string]any{
			"team_id": -1,
			"role":    0,
		}); errTx != nil {
			return errTx
		}
		if errTx := txTeamRepo.DeleteByID(ctx, team.ID); errTx != nil {
			return errTx
		}
		return nil
	})
	if err != nil {
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func hfTeamCreate(ctx *gin.Context) {
	api := &TeamCreateApi{}
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

func hfTeamJoin(ctx *gin.Context) {
	api := &TeamJoinApi{}
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

func hfTeamInfo(ctx *gin.Context) {
	api := &TeamInfoApi{}
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

func hfTeamUpdate(ctx *gin.Context) {
	api := &TeamUpdateApi{}
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

func hfTeamLeave(ctx *gin.Context) {
	api := &TeamLeaveApi{}
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

func hfTeamDisband(ctx *gin.Context) {
	api := &TeamDisbandApi{}
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
