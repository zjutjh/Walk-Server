package user

import (
	"fmt"
	"reflect"
	"runtime"

	"app/comm"
	"app/dao/model"
	"app/dao/repo"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"
)

func NoticeListHandler() gin.HandlerFunc {
	api := NoticeListApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfNoticeList).Pointer()).Name()] = api
	return hfNoticeList
}

type NoticeListApi struct {
	Info     struct{} `name:"未读通知列表" desc:"返回当前用户尚未确认的通知及后端生成的展示文案"`
	Request  struct{}
	Response NoticeListApiResponse
}

type NoticeListApiResponse struct {
	Notices []NoticeItem `json:"notices" desc:"未读通知，按产生时间升序排列"`
}

type NoticeItem struct {
	ID      int64  `json:"id" desc:"通知ID"`
	Content string `json:"content" desc:"通知展示文案"`
}

func (h *NoticeListApi) Run(ctx *gin.Context) kit.Code {
	userID, err := comm.GetUserIDFromCtx(ctx)
	if err != nil || userID <= 0 {
		return comm.CodeNotLoggedIn
	}
	notices, err := repo.NewNoticeRepo().ListUnread(ctx, userID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("查询未读通知失败")
		return comm.CodeServerError
	}
	h.Response.Notices = make([]NoticeItem, 0, len(notices))
	for _, notice := range notices {
		content, ok := noticeContent(notice)
		if !ok {
			nlog.Pick().WithContext(ctx).Warn("忽略未知类型通知")
			continue
		}
		h.Response.Notices = append(h.Response.Notices, NoticeItem{ID: notice.ID, Content: content})
	}
	return comm.CodeOK
}

func noticeContent(notice *model.Notice) (string, bool) {
	switch notice.Type {
	case comm.NoticeTeamPasswordChanged:
		return "队长已修改队伍密码", true
	case comm.NoticeTeamRouteChanged:
		return "队长已修改队伍路线", true
	case comm.NoticeRemovedFromTeam:
		return "你已被移出队伍", true
	case comm.NoticeCaptainTransferred:
		if notice.ActorName == "" {
			return "你已成为新队长", true
		}
		return fmt.Sprintf("%s已将队长移交给你！", notice.ActorName), true
	default:
		return "", false
	}
}

func hfNoticeList(ctx *gin.Context) {
	api := &NoticeListApi{}
	if code := api.Run(ctx); code == comm.CodeOK {
		reply.Reply(ctx, comm.CodeOK, api.Response)
	} else {
		reply.Fail(ctx, code)
	}
}

func NoticeAckHandler() gin.HandlerFunc {
	api := NoticeAckApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfNoticeAck).Pointer()).Name()] = api
	return hfNoticeAck
}

type NoticeAckApi struct {
	Info     struct{} `name:"确认通知" desc:"按通知ID确认当前用户的一条未读通知"`
	Request  NoticeAckApiRequest
	Response struct{}
}

type NoticeAckApiRequest struct {
	Body struct {
		NoticeID int64 `json:"notice_id" desc:"已确认的通知ID" binding:"required,gt=0"`
	}
}

func (h *NoticeAckApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindJSON(&h.Request.Body)
}

func (h *NoticeAckApi) Run(ctx *gin.Context) kit.Code {
	userID, err := comm.GetUserIDFromCtx(ctx)
	if err != nil || userID <= 0 {
		return comm.CodeNotLoggedIn
	}
	if err := repo.NewNoticeRepo().Ack(ctx, userID, h.Request.Body.NoticeID); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("确认通知失败")
		return comm.CodeServerError
	}
	return comm.CodeOK
}

func hfNoticeAck(ctx *gin.Context) {
	api := &NoticeAckApi{}
	if err := api.Init(ctx); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	if code := api.Run(ctx); code == comm.CodeOK {
		reply.Reply(ctx, comm.CodeOK, api.Response)
	} else {
		reply.Fail(ctx, code)
	}
}
