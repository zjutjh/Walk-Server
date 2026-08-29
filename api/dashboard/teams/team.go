package teams

import (
	"encoding/json"
	"errors"
	"reflect"
	"runtime"

	//"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
	"gorm.io/gorm"

	//"fmt"

	"app/comm"
	teamCache "app/dao/cache/team"
	repo "app/dao/repo"
)

// TeamHandler API router注册点
func TeamHandler() gin.HandlerFunc {
	api := TeamApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeam).Pointer()).Name()] = api
	return hfTeam
}

type CaptainInfo struct {
	Phone string `json:"phone" desc:"队长联系电话"`
	Name  string `json:"name" desc:"队长姓名"`
}

type MemberInfo struct {
	Name  string `json:"name" desc:"成员姓名"`
	Phone string `json:"phone" desc:"联系电话"`
	Role  string `json:"role" desc:"人员身份(member成员/captain队长)"`
}

type TeamApi struct {
	Info     struct{}        `name:"获取队伍详细信息" desc:"获取指定队伍的完整详细信息，包括队长和所有队员"`
	Request  TeamApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response TeamApiResponse // API响应数据 (Body中的Data部分)
}

type TeamApiRequest struct {
	Query struct {
		TeamId int `form:"team_id" desc:"队伍ID"`
	}
}

type TeamApiResponse struct {
	TeamId          int64        `json:"team_id" desc:"队伍ID（保留）"`
	Members         []MemberInfo `json:"members" desc:"队员信息列表"`
	LatestPointName string       `json:"latest_point_name" desc:"最新经过点位唯一name"`
	LatestPointTime string       `json:"latest_point_time" desc:"经过点位时间"`
	RouteName       string       `json:"route_name" desc:"路线name"`
	IsLost          bool         `json:"is_lost" desc:"是否失联"`
	IsWrongRoute    bool         `json:"is_wrong_route" desc:"是否走错路线"`
}

// Run Api业务逻辑执行点
func (t *TeamApi) Run(ctx *gin.Context) kit.Code {
	// Redis缓存规划:
	// Key: walk:dashboard:team:by-id:{teamId}
	// Type: String(JSON)
	// TTL: 60s
	teamID := int64(t.Request.Query.TeamId)
	if teamID <= 0 {
		return comm.CodeParameterInvalid
	}
	cached, found, err := teamCache.GetTeamInfo(ctx, teamID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("读取队伍详情缓存失败")
	} else if found {
		cachedResp := TeamApiResponse{}
		err = json.Unmarshal(cached, &cachedResp)
		if err == nil {
			t.Response = cachedResp
			return comm.CodeOK
		}

		nlog.Pick().WithContext(ctx).WithError(err).Warn("解析队伍详情缓存失败")
	}

	teamRepo := repo.NewTeamRepo()
	//fmt.Println("1")
	team, err := teamRepo.GetTeamByID(ctx, teamID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return comm.CodeDataNotFound
	}
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍信息失败")
		return comm.CodeServerError
	}

	members, err := teamRepo.ListTeamMembers(ctx, teamID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍成员失败")
		return comm.CodeServerError
	}

	t.Response.TeamId = team.ID
	t.Response.LatestPointName = team.LatestPointName
	t.Response.RouteName = team.RouteName
	t.Response.IsLost = team.IsLost
	t.Response.IsWrongRoute = team.IsWrongRoute
	if !team.Time.IsZero() {
		t.Response.LatestPointTime = team.Time.UTC().Format("2006-01-02T15:04:05.000Z")
	}

	t.Response.Members = make([]MemberInfo, 0, len(members))
	for _, member := range members {
		t.Response.Members = append(t.Response.Members, MemberInfo{
			Name:  member.Name,
			Phone: member.Phone,
			Role:  normalizeMemberRole(member.Role, member.ID, team.Captain),
		})
	}

	cacheBody, err := json.Marshal(t.Response)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("序列化队伍详情缓存失败")
		return comm.CodeOK
	}

	err = teamCache.SetTeamInfo(ctx, teamID, cacheBody)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("写入队伍详情缓存失败")
	}

	return comm.CodeOK
}

// normalizeMemberRole 兼容历史 role 脏数据，并优先以队长用户 ID 判定 captain。
func normalizeMemberRole(role string, memberID, captainID int64) string {
	if memberID == captainID || strings.EqualFold(role, "captain") {
		return "captain"
	}

	if strings.EqualFold(role, "member") || strings.EqualFold(role, "menber") {
		return "member"
	}

	return role
}

// Init Api初始化 进行参数校验和绑定
func (t *TeamApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindQuery(&t.Request.Query)
}

// hfTeam API执行入口
func hfTeam(ctx *gin.Context) {
	api := &TeamApi{}
	if err := api.Init(ctx); err != nil {
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
