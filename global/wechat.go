package global

import (
	"context"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
)

var (
	MiniProgram     *miniProgram.MiniProgram
	OfficialAccount *officialAccount.OfficialAccount
	Wctx            = context.Background()
)
