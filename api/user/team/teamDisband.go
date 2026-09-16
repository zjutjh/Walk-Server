package team

import (
	"context"
	"errors"
	"reflect"
	"runtime"
	"time"

	teamCache "app/dao/cache/team"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"gorm.io/gorm"

	"app/comm"
)

func TeamDisbandHandler() gin.HandlerFunc {
	api := TeamDisbandApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeamDisband).Pointer()).Name()] = api
	return hfTeamDisband
}

type TeamDisbandApi struct {
	Info     struct{} `name:"解散团队" desc:"队长解散团队"`
	Request  struct{}
	Response struct{}
}

func (h *TeamDisbandApi) Run(ctx *gin.Context) kit.Code {
	if code := comm.CheckBizPhase(comm.PhaseRegistration, comm.PhaseSubmission, comm.PhaseAdjustment); code != comm.CodeOK {
		return code
	}
	person, code := currentTeamUser(ctx)
	if code != comm.CodeOK {
		return code
	}
	if person.Role == comm.RoleUnbind || person.TeamID <= 0 {
		return comm.CodeNotInTeam
	}
	// Read MySQL directly: stale cached submit state must not decide quota refunds.
	teamRepo := repo.NewTeamRepo()
	team, err := teamRepo.GetTeamByID(ctx, person.TeamID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return comm.CodeTeamNotFound
	}
	if err != nil {
		return comm.CodeServerError
	}
	if !(person != nil && team != nil && person.Role == comm.RoleCaptain && team.Captain == person.ID) {
		return comm.CodeNotCaptain
	}
	if team.Submit && !comm.IsInBizPhase(comm.PhaseSubmission, comm.PhaseAdjustment) {
		return comm.CodeTeamSubmitted
	}
	refunded, submittedRoute, submittedDay := false, "", 0
	if team.Submit && comm.IsInBizPhase(comm.PhaseSubmission) {
		var err error
		// Disbanding is allowed throughout submission, including outside daily windows.
		// Require the ledger's original day so a missing record cannot refund the wrong day.
		refunded, submittedRoute, submittedDay, err = teamCache.RollbackTeamSubmit(ctx, team.ID, "", -1)
		if err != nil || !refunded {
			nlog.Pick().WithContext(ctx).WithError(err).Error("解散团队退还名额失败或提交记录缺失")
			return comm.CodeServerError
		}
	}
	if err := teamRepo.DisbandTeam(ctx, team.ID); err != nil {
		if refunded {
			// Compensation must still run if the request context has been cancelled.
			recoveryCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			if restoreErr := teamCache.RestoreSubmittedTeam(recoveryCtx, team.ID, submittedRoute, submittedDay); restoreErr != nil {
				nlog.Pick().WithContext(ctx).WithError(restoreErr).Error("解散团队失败后恢复名额失败，需核对提交记录")
			}
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("解散团队失败")
		return comm.CodeServerError
	}
	if !refunded {
		if err := teamCache.ClearSubmittedTeam(ctx, team.ID); err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("团队已解散，清理提交记录失败")
		}
	}
	_ = teamCache.DelTeamByID(ctx, team.ID)
	_ = teamCache.DeleteTeamInfo(ctx, team.ID)
	if team.Code != "" {
		_ = teamCache.DelTeamIDByCode(ctx, team.Code)
	}
	return comm.CodeOK
}

func hfTeamDisband(ctx *gin.Context) {
	api := &TeamDisbandApi{}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
