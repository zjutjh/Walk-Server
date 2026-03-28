package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"

	"app/comm"
	"app/dao/repo"
)

type RegisterCommonRequest struct {
	Name     string `json:"name" binding:"required"`
	Gender   string `json:"gender" binding:"required" desc:"字符串枚举: male|female"`
	Campus   string `json:"campus" binding:"required" desc:"字符串枚举: zh|pf|mgs"`
	StuID    string `json:"stu_id"`
	Identity string `json:"identity" binding:"required"`
	QQ       string `json:"qq"`
	Wechat   string `json:"wechat"`
	College  string `json:"college" binding:"required"`
	Tel      string `json:"tel" binding:"required"`
}

func doRegister(ctx *gin.Context, req RegisterCommonRequest, personType string) kit.Code {
	gender, ok := comm.ParseGender(req.Gender)
	if !ok {
		return comm.CodeParameterInvalid
	}

	campus, ok := comm.ParseCampus(req.Campus)
	if !ok {
		return comm.CodeParameterInvalid
	}

	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	peopleRepo := repo.NewPeopleRepo()
	existing, err := peopleRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if existing != nil {
		return comm.CodeAlreadyRegistered
	}

	byIdentity, err := peopleRepo.FindByIdentity(ctx, req.Identity)
	if err != nil {
		return comm.CodeDatabaseError
	}
	if byIdentity != nil {
		return comm.CodeAlreadyRegistered
	}

	if personType == comm.MemberTypeStudent && req.StuID != "" {
		byStuID, err := peopleRepo.FindByStuID(ctx, req.StuID)
		if err != nil {
			return comm.CodeDatabaseError
		}
		if byStuID != nil {
			return comm.CodeAlreadyRegistered
		}
	}

	err = peopleRepo.Create(ctx, &repo.Person{
		OpenID:     openID,
		Name:       req.Name,
		Gender:     gender,
		StuID:      req.StuID,
		Campus:     campus,
		Identity:   req.Identity,
		Role:       comm.RoleUnbind,
		QQ:         req.QQ,
		Wechat:     req.Wechat,
		College:    req.College,
		Tel:        req.Tel,
		CreatedOp:  3,
		JoinOp:     5,
		TeamID:     -1,
		Type:       personType,
		WalkStatus: comm.WalkStatusNotStart,
	})
	if err != nil {
		if isDuplicateEntryError(err) {
			return comm.CodeAlreadyRegistered
		}
		nlog.Pick().WithContext(ctx).WithError(err).Error("创建报名记录失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func isDuplicateEntryError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
