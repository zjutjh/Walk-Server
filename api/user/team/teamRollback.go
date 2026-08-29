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

func (h *TeamRollbackApi) Run(ctx *gin.Context) kit.Code {
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	team, code := currentUserTeam(ctx, person)
	if code != comm.CodeOK {
		return code
	}
	if !(person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.ID) {
		return comm.CodeNotCaptain
	}
	if !team.Submit {
		return comm.CodeTeamNotSubmitted
	}
	day, ok := comm.CurrentSubmissionDay()
	if !ok {
		nlog.Pick().WithContext(ctx).Warn("当前阶段不可提交队伍")
		return comm.CodeCannotSubmit
	}
	submitted, submittedDay, err := teamCache.RollbackTeamSubmit(ctx, team.ID, day)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("撤销团队提交归还名额失败")
		return comm.CodeServerError
	}
	if err := repo.NewTeamRepo().UpdateByID(ctx, team.ID, map[string]any{"submit": false}); err != nil {
		if submitted {
			_ = teamCache.RestoreSubmittedTeam(ctx, team.ID, submittedDay)
		}
		return comm.CodeServerError
	}
	if !submitted {
		nlog.Pick().WithContext(ctx).Warn("MySQL 中队伍已提交，但 Redis 中缺少提交名额记录")
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	return comm.CodeOK
}

func hfTeamRollback(ctx *gin.Context) {
	api := &TeamRollbackApi{}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
