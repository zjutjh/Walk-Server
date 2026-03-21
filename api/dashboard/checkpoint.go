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
	"app/middleware"
)

// CheckpointHandler API router注册点
func CheckpointHandler() gin.HandlerFunc {
	api := CheckpointApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfCheckpoint).Pointer()).Name()] = api
	return hfCheckpoint
}

type CheckpointApi struct {
	Info     struct{}              `name:"获取点位详情" desc:"获取指定路线上某个点位的详细信息"`
	Request  CheckpointApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response CheckpointApiResponse // API响应数据 (Body中的Data部分)
}

type CheckpointApiRequest struct {
	Query struct {
		PointName string `form:"point_name" desc:"点位编号，请使用全局唯一name，而不是CPn"`
	}
}

type CheckpointApiResponse struct {
	PassedCount     int `json:"passed_count" desc:"经过该点位的总人数"`
	NotArrivedCount int `json:"not_arrived_count" desc:"未到达该点位的人数"`
}

// Run Api业务逻辑执行点
func (c *CheckpointApi) Run(ctx *gin.Context) kit.Code {
	// Redis缓存规划:
	// Key: walk:dashboard:checkpoint:{campus}:{pointName}
	// Type: String(JSON)
	// TTL: 10s
	admin, ok := middleware.GetAdminInfo(ctx)
	if !ok {
		return comm.CodeUnknownError
	}

	campus := strings.ToLower(strings.TrimSpace(admin.Campus))
	pointName := strings.TrimSpace(c.Request.Query.PointName)
	if campus == "" || pointName == "" {
		return comm.CodeParameterInvalid
	}

	dashboardCache := cachedao.NewDashboardCache()

	// 先走缓存，命中则直接返回。
	cached, found, err := dashboardCache.GetCheckpoint(ctx, campus, pointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("读取点位详情缓存失败")
	} else if found {
		cachedResp := CheckpointApiResponse{}
		err = json.Unmarshal(cached, &cachedResp)
		if err == nil {
			c.Response = cachedResp
			return comm.CodeOK
		}

		nlog.Pick().WithContext(ctx).WithError(err).Warn("解析点位详情缓存失败")
	}

	dashboardRepo := repodao.NewDashboardRepo()
	passedCount, notArrivedCount, err := dashboardRepo.GetCheckpointPeopleCounts(ctx, campus, pointName)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询点位详情统计失败")
		return comm.CodeDatabaseError
	}

	c.Response.PassedCount = int(passedCount)
	c.Response.NotArrivedCount = int(notArrivedCount)

	cacheBody, err := json.Marshal(c.Response)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("序列化点位详情缓存失败")
		return comm.CodeOK
	}

	err = dashboardCache.SetCheckpoint(ctx, campus, pointName, cacheBody)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("写入点位详情缓存失败")
	}

	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (c *CheckpointApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindQuery(&c.Request.Query)
	if err != nil {
		return err
	}
	return err
}

// hfCheckpoint API执行入口
func hfCheckpoint(ctx *gin.Context) {
	api := &CheckpointApi{}
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
