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
	"app/dao/model"
	"app/dao/repo"
)

func UserInfoHandler() gin.HandlerFunc {
	api := UserInfoApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfUserInfo).Pointer()).Name()] = api
	return hfUserInfo
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
	Campus     string `json:"campus" desc:"字符串枚举: zh|pf|mgs"`
	Identity   string `json:"identity"`
	TeamRole   string `json:"team_role" desc:"字符串枚举: unbind|member|captain"`
	QQ         string `json:"qq"`
	Wechat     string `json:"wechat"`
	College    string `json:"college"`
	Tel        string `json:"tel"`
	CreatedOp  uint8  `json:"created_op"`
	JoinOp     uint8  `json:"join_op"`
	TeamID     int64  `json:"team_id"`
	MemberType string `json:"member_type" desc:"字符串枚举: student|teacher|alumnus"`
	WalkStatus string `json:"walk_status" desc:"字符串枚举: notStart|pending|abandoned|inProgress|withdrawn|violated|completed"`
}

type UserInfoTeam struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Num           uint8  `json:"num"`
	Slogan        string `json:"slogan"`
	AllowMatch    bool   `json:"allow_match"`
	Captain       string `json:"captain"`
	RouteName     string `json:"route_name"`
	PrevPointName string `json:"prev_point_name"`
	Submit        bool   `json:"submit"`
	Status        string `json:"status" desc:"字符串枚举: notStart|inProgress|completed|withDrawn"`
	IsLost        bool   `json:"is_lost"`
}

func (h *UserInfoApi) Init(ctx *gin.Context) (err error) {
	return err
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

func toUserInfoPerson(person *model.People) *UserInfoPerson {
	if person == nil {
		return nil
	}
	return &UserInfoPerson{
		ID:         person.ID,
		Name:       person.Name,
		Gender:     comm.FormatGender(person.Gender),
		StuID:      person.StuID,
		Campus:     person.Campus,
		Identity:   person.Identity,
		TeamRole:   person.Role,
		QQ:         person.Qq,
		Wechat:     person.Wechat,
		College:    person.College,
		Tel:        person.Tel,
		CreatedOp:  uint8(person.CreatedOp),
		JoinOp:     uint8(person.JoinOp),
		TeamID:     person.TeamID,
		MemberType: person.Type,
		WalkStatus: person.WalkStatus,
	}
}

func toUserInfoTeam(team *model.Team) *UserInfoTeam {
	if team == nil {
		return nil
	}
	return &UserInfoTeam{
		ID:            team.ID,
		Name:          team.Name,
		Num:           uint8(team.Num),
		Slogan:        team.Slogan,
		AllowMatch:    team.AllowMatch != 0,
		Captain:       team.Captain,
		RouteName:     team.RouteName,
		PrevPointName: team.PrevPointName,
		Submit:        team.Submit != 0,
		Status:        team.Status,
		IsLost:        team.IsLost != 0,
	}
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
