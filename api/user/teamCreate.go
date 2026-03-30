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
	"app/dao/model"
	"app/dao/query"
	"app/dao/repo"
)

func TeamCreateHandler() gin.HandlerFunc {
	api := TeamCreateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamCreate).Pointer()).Name()] = api
	return hfTeamCreate
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
	RouteName  string `json:"route_name"`
}

type TeamCreateApiResponse struct {
	TeamID int64 `json:"team_id"`
}

func (h *TeamCreateApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *TeamCreateApi) Run(ctx *gin.Context) kit.Code {
	teamName := strings.TrimSpace(h.Request.Name)
	if teamName == "" {
		return comm.CodeParameterInvalid
	}

	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	teamRepo := repo.NewTeamRepo()

	person, err := peopleRepo.FindPeopleByOpenID(ctx, openID)
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

	duplicated, err := teamRepo.FindTeamByName(ctx, teamName)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if duplicated != nil {
		return comm.CodeTeamNameDuplicated
	}

	err = query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)
		txTeamRepo := repo.NewTeamRepoWithTx(tx)

		team := &model.Team{
			Name:       teamName,
			Num:        1,
			Password:   h.Request.Password,
			Slogan:     h.Request.Slogan,
			AllowMatch: boolToInt8(h.Request.AllowMatch),
			Captain:    openID,
			RouteName:  h.Request.RouteName,
			Submit:     0,
			Status:     comm.TeamStatusNotStart,
			Code:       "",
			IsLost:     0,
		}
		if errTx := txTeamRepo.Create(ctx, team); errTx != nil {
			if isDuplicateEntryError(errTx) {
				return errTeamNameDuplicated
			}
			return errTx
		}

		if errTx := txPeopleRepo.UpdateByOpenID(ctx, openID, map[string]any{
			"team_id":     team.ID,
			"role":        comm.RoleCaptain,
			"created_op":  person.CreatedOp - 1,
			"walk_status": comm.WalkStatusNotStart,
		}); errTx != nil {
			return errTx
		}

		h.Response.TeamID = team.ID
		return nil
	})
	if err != nil {
		if errors.Is(err, errTeamNameDuplicated) {
			return comm.CodeTeamNameDuplicated
		}
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
