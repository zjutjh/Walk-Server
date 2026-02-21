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
)

func GetTeamStatusHandler() gin.HandlerFunc {
    api := GetTeamStatusApi{}
    swagger.CM[runtime.FuncForPC(reflect.ValueOf(getTeamStatus).Pointer()).Name()] = api
    return getTeamStatus
}

type GetTeamStatusApi struct {
    Info     struct{} `name:"获取团队状态"`
    Request  GetTeamStatusApiRequest
    Response GetTeamStatusApiResponse
}

type GetTeamStatusApiRequest struct {
}

type GetTeamStatusApiResponse struct {
	Admin AdminResponse `json:"admin"`
	Team TeamResponse `json:"team"`
	Member MemberResponse `json:"member"`
}

type AdminResponse struct{
	AdminID int `json:"admin_id"`
	Name string `json:"name"`
	Account string `json:"account"`
	Point string `json:"point"`
	RouteCode string `json:"route_code"`
}

type TeamResponse struct{
	TeamID int `json:"team_id"`
	Name string `json:"name"`
	Point string `json:"point"`
	RouteCode string `json:"route_code"`
	Slogan string `json:"slogan"`
	ProgressNum int `json:"progress_num"`
	Status string `json:"status" decs:"有未出发,进行中,已完成,已下撤"`
}

type MemberResponse struct{
	Campus int `json:"campus"`
	Gender int `json:"gender"`
	Name string `json:"name"`
	OpenID string `json:"open_id"`
	WalkStatus string `json:"walk_status"`
	Contact ContactResponse `json:"contact"`
	Type int `json:"type"`
}

type ContactResponse struct{
	QQ string `json:"qq"`
	Wechat string `json:"wechat"`
	Tel string `json:"tel"`
}

// Run Api业务逻辑执行点
func (h *GetTeamStatusApi) Run(ctx *gin.Context) kit.Code {
    return comm.CodeOK
}

// Run Api初始化 进行参数校验和绑定
func (h *GetTeamStatusApi) Init(ctx *gin.Context) (err error) {
    return err
}

func getTeamStatus(ctx *gin.Context) {
    api := &GetTeamStatusApi{}
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