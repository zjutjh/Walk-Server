package api

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	"app/dao/repo"
)

func UserInfoHandler() gin.HandlerFunc {
	api := UserInfoApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUserInfo).Pointer()).Name()] = api
	return hfUserInfo
}

func UserModifyHandler() gin.HandlerFunc {
	api := UserModifyApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUserModify).Pointer()).Name()] = api
	return hfUserModify
}

type UserInfoApi struct {
	Info     struct{} `name:"用户信息" desc:"获取当前登录用户信息"`
	Request  struct{}
	Response UserInfoApiResponse
}

type UserInfoApiResponse struct {
	Person *UserInfoPerson `json:"person"`
	Team   *UserInfoTeam   `json:"team"`
}

type UserInfoPerson struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Gender     string `json:"gender" desc:"字符串枚举: male|female"`
	StuID      string `json:"stu_id"`
	Campus     string `json:"campus" desc:"字符串枚举: chaohui|pingfeng|moganshan"`
	Identity   string `json:"identity"`
	TeamRole   string `json:"team_role" desc:"字符串枚举: none|member|captain"`
	QQ         string `json:"qq"`
	Wechat     string `json:"wechat"`
	College    string `json:"college"`
	Tel        string `json:"tel"`
	CreatedOp  uint8  `json:"created_op"`
	JoinOp     uint8  `json:"join_op"`
	TeamID     int64  `json:"team_id"`
	MemberType string `json:"member_type" desc:"字符串枚举: student|teacher|alumnus"`
	WalkStatus string `json:"walk_status" desc:"字符串枚举: not_started|ready|in_progress|quit|withdrawn|violation|finished"`
}

type UserInfoTeam struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Num        uint8  `json:"num"`
	Slogan     string `json:"slogan"`
	AllowMatch bool   `json:"allow_match"`
	Captain    string `json:"captain"`
	RouteID    int64  `json:"route_id"`
	PointID    int8   `json:"point_id"`
	Submit     bool   `json:"submit"`
	Status     string `json:"status" desc:"字符串枚举: not_started|in_progress|finished|withdrawn"`
	IsLost     bool   `json:"is_lost"`
}

type UserModifyApi struct {
	Info     struct{} `name:"修改用户信息" desc:"修改当前登录用户可编辑信息"`
	Request  UserModifyApiRequest
	Response struct{}
}

type UserModifyApiRequest struct {
	QQ     string `json:"qq"`
	Wechat string `json:"wechat"`
	Tel    string `json:"tel"`
}

func (h *UserInfoApi) Init(ctx *gin.Context) (err error) {
	return err
}

func (h *UserModifyApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request)
}

func (h *UserInfoApi) Run(ctx *gin.Context) kit.Code {
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

	h.Response.Person = toUserInfoPerson(person)
	if person.TeamID > 0 {
		team, err := teamRepo.FindByID(ctx, person.TeamID)
		if err != nil {
			return comm.CodeDatabaseError
		}
		h.Response.Team = toUserInfoTeam(team)
	}

	return comm.CodeOK
}

func toUserInfoPerson(person *repo.Person) *UserInfoPerson {
	if person == nil {
		return nil
	}
	return &UserInfoPerson{
		ID:         person.ID,
		Name:       person.Name,
		Gender:     formatGender(person.Gender),
		StuID:      person.StuID,
		Campus:     formatCampus(person.Campus),
		Identity:   person.Identity,
		TeamRole:   formatTeamRole(person.Role),
		QQ:         person.QQ,
		Wechat:     person.Wechat,
		College:    person.College,
		Tel:        person.Tel,
		CreatedOp:  person.CreatedOp,
		JoinOp:     person.JoinOp,
		TeamID:     person.TeamID,
		MemberType: formatMemberType(person.Type),
		WalkStatus: formatWalkStatus(person.WalkStatus),
	}
}

func toUserInfoTeam(team *repo.Team) *UserInfoTeam {
	if team == nil {
		return nil
	}
	return &UserInfoTeam{
		ID:         team.ID,
		Name:       team.Name,
		Num:        team.Num,
		Slogan:     team.Slogan,
		AllowMatch: team.AllowMatch,
		Captain:    team.Captain,
		RouteID:    team.RouteID,
		PointID:    team.PointID,
		Submit:     team.Submit,
		Status:     formatTeamStatus(team.Status),
		IsLost:     team.IsLost,
	}
}

func (h *UserModifyApi) Run(ctx *gin.Context) kit.Code {
	openID := comm.GetOpenIDFromCtx(ctx)
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	err := repo.NewPeopleRepo().UpdateByOpenID(ctx, openID, map[string]any{
		"qq":     h.Request.QQ,
		"wechat": h.Request.Wechat,
		"tel":    h.Request.Tel,
	})
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("更新用户信息失败")
		return comm.CodeDatabaseError
	}

	return comm.CodeOK
}

func hfUserInfo(ctx *gin.Context) {
	api := &UserInfoApi{}
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

func hfUserModify(ctx *gin.Context) {
	api := &UserModifyApi{}
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
