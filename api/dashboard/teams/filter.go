package teams

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	teamCache "app/dao/cache/team"
	repodao "app/dao/repo/dashboard"
)

// FilterHandler API router注册点
func FilterHandler() gin.HandlerFunc {
	api := FilterApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfFilter).Pointer()).Name()] = api
	return hfFilter
}

type FilterApi struct {
	Info     struct{}          `name:"筛选队伍" desc:"搜索队伍和获取指定路段上的队伍列表合并接口，按更新时间正序排序（距上次更新时间最长的在最前面）\ncampus必填，作为第一道筛选\nkey和toPointName不可同时为空"`
	Request  FilterApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response FilterApiResponse // API响应数据 (Body中的Data部分)
}

type FilterApiRequest struct {
	Query struct {
		Campus        string `form:"campus" desc:"校区（必填 pf/mgs）"`
		ToPointName   string `form:"to_point_name" desc:"结束点位name，全局唯一，不是CPn"`
		PrevPointName string `form:"prev_point_name" desc:"上一点位name，合流点一定要给"`
		Key           string `form:"key" desc:"搜索关键词"`
		SearchType    string `form:"search_type" desc:"搜索类型（team_id/captain_phone/captain_name）"`
		Limit         int    `form:"limit" desc:"返回数量"`
		Cursor        int    `form:"cursor" desc:"指针，从0开始"`
	}
}

type FilterApiResponse struct {
	TotalCount int             `json:"total_count" desc:"满足要求的总队伍数"`
	NextCursor int             `json:"next_cursor" desc:"下一页游标，为0则表示无更多数据"`
	Teams      []TeamBriefInfo `json:"teams" desc:"队伍列表"`
}

type TeamBriefInfo struct {
	TeamId        string `json:"team_id" desc:"队伍ID"`
	CaptainName   string `json:"captain_name" desc:"队长姓名"`
	CaptainPhone  string `json:"captain_phone" desc:"队长联系电话"`
	PrevPointName string `json:"prev_point_name" desc:"最新经过点位唯一name"`
	PrevPointTime string `json:"prev_point_time" desc:"最新经过点位时间"`
	RouteName     string `json:"route_name" desc:"路线name"`
	IsLost        bool   `json:"is_lost" desc:"是否失联"`
}

// Run Api业务逻辑执行点
func (f *FilterApi) Run(ctx *gin.Context) kit.Code {
	// Redis缓存规划:
	// Key: walk:dashboard:teams:filter:{campus}:{queryHash}
	// Type: String(JSON)
	// TTL: 20~30s
	campus := strings.ToLower(strings.TrimSpace(f.Request.Query.Campus))
	toPointName := strings.TrimSpace(f.Request.Query.ToPointName)
	prevPointName := strings.TrimSpace(f.Request.Query.PrevPointName)
	key := strings.TrimSpace(f.Request.Query.Key)
	searchType := strings.ToLower(strings.TrimSpace(f.Request.Query.SearchType))

	// 参数校验
	if campus == "" {
		return comm.CodeInsufficientParams
	}
	if key == "" && toPointName == "" {
		return comm.CodeInsufficientParams
	}
	if toPointName == "" && prevPointName != "" {
		return comm.CodeInsufficientParams
	}

	if key != "" {
		switch searchType {
		case "team_id", "captain_phone", "captain_name":
		default:
			return comm.CodeParameterInvalid
		}

		if searchType == "team_id" {
			teamID, err := strconv.ParseInt(key, 10, 64)
			if err != nil || teamID <= 0 {
				return comm.CodeParameterInvalid
			}
		}
	} else if searchType != "" {
		return comm.CodeParameterInvalid
	}

	
	limit := f.Request.Query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	cursor := f.Request.Query.Cursor
	if cursor < 0 {
		cursor = 0
	}

	filterQuery := repodao.TeamFilterQuery{
		Campus:        campus,
		ToPointName:   toPointName,
		PrevPointName: prevPointName,
		Key:           key,
		SearchType:    searchType,
		Limit:         limit,
		Offset:        cursor,
	}

	queryHash := buildFilterQueryHash(filterQuery)

	// 先走缓存，命中则直接返回。
	cached, found, err := teamCache.GetTeamFilter(ctx, campus, queryHash)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("读取队伍筛选缓存失败")
	} else if found {
		cachedResp := FilterApiResponse{}
		err = json.Unmarshal(cached, &cachedResp)
		if err == nil {
			f.Response = cachedResp
			return comm.CodeOK
		}

		nlog.Pick().WithContext(ctx).WithError(err).Warn("解析队伍筛选缓存失败")
	}

	dashboardRepo := repodao.NewDashboardRepo()

	totalCount, err := dashboardRepo.CountTeamsByFilter(ctx, filterQuery)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("统计队伍筛选结果失败")
		return comm.CodeDatabaseError
	}

	teams, err := dashboardRepo.ListTeamsByFilter(ctx, filterQuery)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询队伍筛选列表失败")
		return comm.CodeDatabaseError
	}

	f.Response.TotalCount = int(totalCount)
	f.Response.NextCursor = 0
	f.Response.Teams = make([]TeamBriefInfo, 0, len(teams))

	for _, team := range teams {
		item := TeamBriefInfo{
			TeamId:        strconv.FormatInt(team.TeamID, 10),
			CaptainName:   team.CaptainName,
			CaptainPhone:  team.CaptainPhone,
			PrevPointName: team.PrevPointName,
			RouteName:     team.RouteName,
			IsLost:        team.IsLost == 1,
		}
		if team.PrevPointTime.Valid {
			item.PrevPointTime = team.PrevPointTime.Time.UTC().Format("2006-01-02T15:04:05.000Z")
		}

		f.Response.Teams = append(f.Response.Teams, item)
	}

	nextOffset := cursor + len(f.Response.Teams)
	if int64(nextOffset) < totalCount {
		f.Response.NextCursor = nextOffset
	}

	cacheBody, err := json.Marshal(f.Response)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("序列化队伍筛选缓存失败")
		return comm.CodeOK
	}

	err = teamCache.SetTeamFilter(ctx, campus, queryHash, cacheBody)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("写入队伍筛选缓存失败")
	}

	return comm.CodeOK
}

func buildFilterQueryHash(query repodao.TeamFilterQuery) string {
	raw := query.Campus + "|" + query.ToPointName + "|" + query.PrevPointName + "|" + query.Key + "|" + query.SearchType + "|" + strconv.Itoa(query.Limit) + "|" + strconv.Itoa(query.Offset)
	hash := sha1.Sum([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// Init Api初始化 进行参数校验和绑定
func (f *FilterApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&f.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfFilter API执行入口
func hfFilter(ctx *gin.Context) {
	api := &FilterApi{}
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
