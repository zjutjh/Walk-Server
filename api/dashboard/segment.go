package dashboard

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
	cachedao "app/dao/cache/dashboard"
	repodao "app/dao/repo/dashboard"
	repo "app/dao/repo/admin"
)

// SegmentHandler API router注册点
func SegmentHandler() gin.HandlerFunc {
	api := SegmentApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfSegment).Pointer()).Name()] = api
	return hfSegment
}

type SegmentApi struct {
	Info     struct{}           `name:"获取路段（边）信息" desc:"获取指定路段的人数信息"`
	Request  SegmentApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response SegmentApiResponse // API响应数据 (Body中的Data部分)
}

type SegmentApiRequest struct {
	Query struct {
		ToPointName   string `form:"to_point_name" desc:"结束点位name，全局唯一，不是CPn"`
		PrevPointName string `form:"prev_point_name" desc:"起始点位name，合流点一定要给"`
	}
}

type SegmentApiResponse struct {
	Number int `json:"number" desc:"该路段上的人数"`
}

// Run Api业务逻辑执行点
func (s *SegmentApi) Run(ctx *gin.Context) kit.Code {
	// Redis缓存规划:
	// Key: walk:dashboard:segment:{campus}:{prevPoint}:{toPoint}
	// Type: String(JSON)
	// TTL: 15s
	admin, ok := repo.GetAdminInfo(ctx)
	if !ok {
		return comm.CodeUnknownError
	}

	campus := strings.ToLower(strings.TrimSpace(admin.Campus))
	prevPointName := strings.TrimSpace(s.Request.Query.PrevPointName)
	toPointName := strings.TrimSpace(s.Request.Query.ToPointName)
	if campus == "" || prevPointName == "" || toPointName == "" {
		return comm.CodeParameterInvalid
	}

	dashboardCache := cachedao.NewDashboardCache()

	// 先走缓存，命中则直接返回。
	cached, found, err := dashboardCache.GetSegment(ctx, campus, prevPointName, toPointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("读取路段人数缓存失败")
	} else if found {
		cachedResp := SegmentApiResponse{}
		err = json.Unmarshal(cached, &cachedResp)
		if err == nil {
			s.Response = cachedResp
			return comm.CodeOK
		}

		nlog.Pick().WithContext(ctx).WithError(err).Warn("解析路段人数缓存失败")
	}

	dashboardRepo := repodao.NewDashboardRepo()
	peopleCount, err := dashboardRepo.CountPeopleOnSegment(ctx, campus, prevPointName, toPointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询路段人数失败")
		return comm.CodeDatabaseError
	}

	s.Response.Number = int(peopleCount)

	cacheBody, err := json.Marshal(s.Response)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("序列化路段人数缓存失败")
		return comm.CodeOK
	}

	err = dashboardCache.SetSegment(ctx, campus, prevPointName, toPointName, cacheBody)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("写入路段人数缓存失败")
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (s *SegmentApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&s.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfSegment API执行入口
func hfSegment(ctx *gin.Context) {
	api := &SegmentApi{}
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
