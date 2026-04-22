package api

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/repo"
)

func TeamUpdateHandler() gin.HandlerFunc {
	api := TeamUpdateApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamUpdate).Pointer()).Name()] = api
	return hfTeamUpdate
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
	RouteName  string `json:"route_name"`
}

func (h *TeamUpdateApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *TeamUpdateApi) Run(ctx *gin.Context) kit.Code {
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
	if person == nil || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}
	if person.Role != comm.RoleCaptain {
		return comm.CodeNotCaptain
	}

	team, err := teamRepo.FindTeamByID(ctx, person.TeamID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if team == nil {
		return comm.CodeDataNotFound
	}
	if team.Submit != 0 {
		return comm.CodeTeamSubmitted
	}

	duplicated, err := teamRepo.FindByNameExceptID(ctx, teamName, team.ID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if duplicated != nil {
		return comm.CodeTeamNameDuplicated
	}

	hashedPassword, err := hashTeamPassword(h.Request.Password)
	if err != nil {
		return comm.CodeUnknownError
	}

	err = teamRepo.UpdateByID(ctx, team.ID, map[string]any{
		"name":        teamName,
		"password":    hashedPassword,
		"slogan":      h.Request.Slogan,
		"allow_match": boolToInt8(h.Request.AllowMatch),
		"route_name":  h.Request.RouteName,
	})
	if err != nil {
		if isDuplicateEntryError(err) {
			return comm.CodeTeamNameDuplicated
		}
		return comm.CodeDatabaseError
	}
	return comm.CodeOK
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
