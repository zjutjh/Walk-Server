package team

import (
	"reflect"
	"runtime"

	teamCache "app/dao/cache/team"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func TeamRollbackHandler() gin.HandlerFunc {
	api := TeamRollbackApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamRollback).Pointer()).Name()] = api
	return hfTeamRollback
}

type TeamRollbackApi struct {
	Info     struct{} `name:"撤销提交" `
	Request  struct{}
	Response struct{}
}

func (h *TeamRollbackApi) Init(ctx *gin.Context) error { return nil }
func (h *TeamRollbackApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	if !(person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.OpenID) {
		return comm.CodeNotCaptain
	}
	day, routeCode, code := teamQuotaRoute(ctx, team.RouteName)
	if code != comm.CodeOK {
		return code
	}
	submitted, err := teamCache.RollbackTeamSubmit(ctx, team.ID, day, routeCode)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("撤销团队提交归还名额失败")
		return comm.CodeServerError
	}
	if !submitted {
		return comm.CodeTeamNotSubmitted
	}
	if err := repo.NewTeamRepo().UpdateByID(ctx, team.ID, map[string]any{"submit": false}); err != nil {
		_ = teamCache.RestoreSubmittedTeam(ctx, team.ID, day, routeCode)
		return comm.CodeServerError
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	return comm.CodeOK
}

func hfTeamRollback(ctx *gin.Context) {
	api := &TeamRollbackApi{}
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
