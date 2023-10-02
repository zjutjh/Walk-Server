package global

import (
	"context"
	"github.com/silenceper/wechat/v2/miniprogram"
)

var (
	MiniProgram *miniprogram.MiniProgram
	Wctx        = context.Background()
)
