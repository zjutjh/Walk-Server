package comm

import (
	"fmt"
	"time"
)

func ValidateBizConfig() error {
	periods := []struct {
		name   string
		period TimeRangeConfig
	}{
		{"registration", BizConf.Phases.Registration},
		{"submission", BizConf.Phases.Submission.TimeRangeConfig},
		{"adjustment", BizConf.Phases.Adjustment},
		{"preparation", BizConf.Phases.Preparation},
		{"activity", BizConf.Phases.Activity},
	}

	for _, item := range periods {
		if err := validateTimeRange("biz.phases."+item.name, item.period); err != nil {
			return err
		}
	}
	for i := 1; i < len(periods); i++ {
		previousEnd, _ := parseBizTime(periods[i-1].period.End)
		currentStart, _ := parseBizTime(periods[i].period.Start)
		if previousEnd.After(currentStart) {
			return fmt.Errorf("biz phases must not overlap and must follow configured order")
		}
	}
	if err := validateSubmissionConfig(); err != nil {
		return err
	}
	if len(BizConf.IdentitySecret) < 32 {
		return fmt.Errorf("biz.identity_secret must be at least 32 characters")
	}
	if BizConf.TeamTotalLimit <= 0 {
		return fmt.Errorf("biz.team_total_limit must be greater than 0")
	}
	return nil
}

func validateSubmissionConfig() error {
	phase := BizConf.Phases.Submission
	dailyStart, err := time.Parse("15:04:05", phase.DailyStartTime)
	if err != nil {
		return fmt.Errorf("biz.phases.submission.daily_start_time must use format %q: %w", "15:04:05", err)
	}
	dailyEnd, err := time.Parse("15:04:05", phase.DailyEndTime)
	if err != nil {
		return fmt.Errorf("biz.phases.submission.daily_end_time must use format %q: %w", "15:04:05", err)
	}
	if !dailyStart.Before(dailyEnd) {
		return fmt.Errorf("biz.phases.submission.daily_start_time must be before daily_end_time")
	}
	if len(BizConf.DailyTeamLimits) == 0 {
		return fmt.Errorf("biz.daily_team_limits must not be empty")
	}
	for day, limit := range BizConf.DailyTeamLimits {
		if limit < 0 {
			return fmt.Errorf("biz.daily_team_limits[%d] must not be negative", day)
		}
	}
	start, _ := parseBizTime(phase.Start)
	end, _ := parseBizTime(phase.End)
	wantDays := daysBetween(start, end) + 1
	if len(BizConf.DailyTeamLimits) != wantDays {
		return fmt.Errorf("biz.daily_team_limits length must equal submission phase days: want %d", wantDays)
	}
	return nil
}

func validateTimeRange(name string, period TimeRangeConfig) error {
	start, err := parseBizTime(period.Start)
	if err != nil {
		return fmt.Errorf("%s.start must use format %q: %w", name, time.DateTime, err)
	}
	end, err := parseBizTime(period.End)
	if err != nil {
		return fmt.Errorf("%s.end must use format %q: %w", name, time.DateTime, err)
	}
	if !start.Before(end) {
		return fmt.Errorf("%s.start must be before end", name)
	}
	return nil
}
