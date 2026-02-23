package dashboard

import (
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

// TeamHandler API router注册点
func TeamHandler() gin.HandlerFunc {
	api := TeamApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfTeam).Pointer()).Name()] = api
	return hfTeam
}

type CaptainInfo struct {
	Phone string `json:"phone" desc:"队长联系电话"`
}

type MemberInfo struct {
	Index int    `json:"index" desc:"成员序号"`
	Phone string `json:"phone" desc:"联系电话"`
}

type TeamApi struct {
	Info     struct{}        `name:"获取队伍详细信息" desc:"获取指定队伍的完整详细信息，包括所有队员"`
	Request  TeamApiRequest  // API请求参数 (Uri/Header/Query/Body)
	Response TeamApiResponse // API响应数据 (Body中的Data部分)
}

type TeamApiRequest struct {
	Uri struct {
		TeamId string `uri:"team_id" desc:"队伍ID"`
	}
}

type TeamApiResponse struct {
	TeamId               int          `json:"team_id" desc:"队伍ID"`
	RouteCode            string       `json:"route_code" desc:"路线ID"`
	Captain              CaptainInfo  `json:"captain" desc:"队长信息"`
	Members              []MemberInfo `json:"members" desc:"队员信息列表（不包括队长）"`
	LatestCheckpointId   string       `json:"latest_checkpoint_id" desc:"最新经过点位唯一id"`
	LatestCheckpointTime string       `json:"latest_checkpoint_time" desc:"经过点位时间"`
}

// Run Api业务逻辑执行点
func (t *TeamApi) Run(ctx *gin.Context) kit.Code {
	// TODO: 在此处编写接口业务逻辑
	return comm.CodeOK
}

// Init Api初始化 进行参数校验和绑定
func (t *TeamApi) Init(ctx *gin.Context) (err error) {
	err = ctx.ShouldBindUri(&t.Request.Uri)
	if err != nil {
		return err
	}
	return err
}

// hfTeam API执行入口
func hfTeam(ctx *gin.Context) {
	api := &TeamApi{}
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
