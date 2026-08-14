package register

import (
	"context"
	"fmt"

	"github.com/zjutjh/mygo/config"
	"github.com/zjutjh/mygo/feishu"
	"github.com/zjutjh/mygo/foundation/kernel"
	"github.com/zjutjh/mygo/jwt"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/lock"
	"github.com/zjutjh/mygo/ndb"
	"github.com/zjutjh/mygo/nedis"
	"github.com/zjutjh/mygo/nesty"
	"github.com/zjutjh/mygo/nlog"

	"app/comm"
	teamCache "app/dao/cache/team"
	"app/register/generate"
)

func Boot() kernel.BootList {
	return kernel.BootList{
		// 基础引导器
		feishu.Boot(),   // 飞书Bot (消息提醒)
		nlog.Boot(),     // 业务日志
		generate.Boot(), // 导入生成代码

		// Client引导器
		ndb.Boot(),   // DB
		nedis.Boot(), // Redis
		jwt.Boot[string](),
		lock.Boot(),  // Redis Lock
		nesty.Boot(), // HTTP Client

		// 业务引导器
		BizConfBoot(),
		TeamQuotaBoot(),
		AppBoot(),
	}
}

func GenerateBoot() kernel.BootList {
	return kernel.BootList{
		feishu.Boot(), // 飞书Bot (消息提醒)
		nlog.Boot(),   // 业务日志
		ndb.Boot(),    // DB
	}
}

// BizConfBoot 初始化应用业务配置引导器
func BizConfBoot() func() error {
	return func() error {
		err := config.Pick().UnmarshalKey("biz", &comm.BizConf)
		if err != nil {
			return fmt.Errorf("%w: 解析应用业务配置错误: %w", kit.ErrDataUnmarshal, err)
		}
		legacy := struct {
			StartDate   string `mapstructure:"startDate"`
			ExpiredDate string `mapstructure:"expiredDate"`
		}{}
		if err := config.Pick().Unmarshal(&legacy); err != nil {
			return fmt.Errorf("%w: 解析旧版业务配置错误: %w", kit.ErrDataUnmarshal, err)
		}
		if comm.BizConf.StartDate == "" {
			comm.BizConf.StartDate = legacy.StartDate
		}
		if comm.BizConf.ExpiredDate == "" {
			comm.BizConf.ExpiredDate = legacy.ExpiredDate
		}
		if err := comm.ValidateBizConfig(); err != nil {
			return fmt.Errorf("业务配置校验失败: %w", err)
		}
		return nil
	}
}

// AppBoot 应用定制引导器
func AppBoot() func() error {
	return func() error {
		return nil
	}
}

func TeamQuotaBoot() func() error {
	return func() error {
		if err := teamCache.InitTotalTeamQuota(context.Background(), comm.BizConf.TeamTotalLimit); err != nil {
			return err
		}
		for day, limit := range comm.BizConf.DailyTeamLimits {
			if err := teamCache.InitDailyTeamQuota(context.Background(), day, limit); err != nil {
				return err
			}
		}
		return nil
	}
}
