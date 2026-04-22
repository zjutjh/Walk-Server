package api

import (
	"errors"
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
	"app/dao/query"
	repo "app/dao/repo"
)

func UpdateUserHandler() gin.HandlerFunc {
	api := UpdateUserApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(updateUser).Pointer()).Name()] = api
	return updateUser
}

type UpdateUserApi struct {
	Info     struct{} `name:"更改人员状态"`
	Request  UpdateUserApiRequest
	Response UpdateUserApiResponse
}

type UpdateUserApiRequest struct {
	Body struct {
		UserID int    `json:"user_id" desc:"用户编号" binding:"required"`
		Status string `json:"status" desc:"未开始,待出发,已放弃,进行中,已下撤,已违规,已完成" binding:"required"`
	}
}

type UpdateUserApiResponse struct {
}

func (u *UpdateUserApi) Run(ctx *gin.Context) kit.Code {
	peopleRepo := repo.NewPeopleRepo()

	user, err := peopleRepo.FindPeopleByID(ctx, int64(u.Request.Body.UserID))
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询人员失败")
		return comm.CodeDatabaseError
	}
	if user == nil {
		return comm.CodePeopleNotFound
	}

	if user.TeamID > 0 {
		mutex := comm.NewTeamMutex(user.TeamID)
		if err := mutex.Lock(); err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("获取队伍成员状态更新锁失败")
			return comm.CodeTooFrequently
		}
		defer func() {
			if _, err := mutex.Unlock(); err != nil {
				nlog.Pick().WithContext(ctx).WithError(err).Warn("释放队伍成员状态更新锁失败")
			}
		}()
	}

	err = query.Use(ndb.Pick()).Transaction(func(tx *query.Query) error {
		txTeamRepo := repo.NewTeamRepoWithTx(tx)
		txPeopleRepo := repo.NewPeopleRepoWithTx(tx)

		if err := txPeopleRepo.UpdateWalkStatus(ctx, user.ID, u.Request.Body.Status); err != nil {
			return err
		}

		team, err := txTeamRepo.FindTeamByID(ctx, user.TeamID)
		if err != nil {
			return err
		}
		if team == nil {
			return gorm.ErrRecordNotFound
		}

		if team.Status == comm.TeamStatusNotStart {
			memberCount, err := txPeopleRepo.CountMembersByTeamID(ctx, user.TeamID)
			if err != nil {
				return err
			}
			abandonedCount, err := txPeopleRepo.CountMembersByStatus(ctx, user.TeamID, comm.WalkStatusAbandoned)
			if err != nil {
				return err
			}
			if memberCount > 0 && memberCount == abandonedCount {
				return txTeamRepo.UpdateByID(ctx, user.TeamID, map[string]any{"status": comm.TeamStatusCompleted})
			}
			return nil
		}

		inProgressCount, err := txPeopleRepo.CountMembersByStatus(ctx, user.TeamID, comm.WalkStatusInProgress)
		if err != nil {
			return err
		}
		if inProgressCount > 0 {
			if team.Status != comm.TeamStatusInProgress {
				return txTeamRepo.UpdateByID(ctx, user.TeamID, map[string]any{"status": comm.TeamStatusInProgress})
			}
			return nil
		}

		if u.Request.Body.Status != comm.WalkStatusWithdrawn {
			return txTeamRepo.UpdateByID(ctx, user.TeamID, map[string]any{"status": comm.TeamStatusCompleted})
		}

		if team.Status != comm.TeamStatusCompleted {
			return txTeamRepo.UpdateByID(ctx, user.TeamID, map[string]any{"status": comm.TeamStatusWithdrawn})
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return comm.CodeDataNotFound
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("更改人员状态失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func (u *UpdateUserApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindJSON(&u.Request.Body)
	if err != nil {
		return err
	}
	return err
}

func updateUser(ctx *gin.Context) {
	api := &UpdateUserApi{}
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
