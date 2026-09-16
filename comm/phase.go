package comm

import (
	"time"

	"github.com/zjutjh/mygo/kit"
)

func IsInBizPhase(allowed ...BizPhase) bool {
	now := time.Now()
	for _, phase := range allowed {
		if period, ok := phaseTimeRange(phase); ok && inTimeRange(now, period) {
			return true
		}
	}
	return false
}

// CurrentBizPhase 返回当前业务时期；不在任何已配置时期内时返回空字符串。
func CurrentBizPhase() BizPhase {
	return CurrentBizPhaseAt(time.Now())
}

func CurrentBizPhaseAt(now time.Time) BizPhase {
	for _, phase := range []BizPhase{
		PhaseRegistration,
		PhaseSubmission,
		PhaseAdjustment,
		PhasePreparation,
		PhaseActivity,
	} {
		if period, ok := phaseTimeRange(phase); ok && inTimeRange(now, period) {
			return phase
		}
	}
	return ""
}

func CheckBizPhase(allowed ...BizPhase) kit.Code {
	if IsInBizPhase(allowed...) {
		return CodeOK
	}
	if IsInBizPhase(PhasePreparation) {
		return CodePreparationForbidden
	}
	if IsInBizPhase(PhaseActivity) {
		return CodeActivityForbidden
	}
	return CodePermissionDenied
}

func CurrentSubmissionDay() (int, bool) {
	now := time.Now()
	phase := BizConf.Phases.Submission
	if !inTimeRange(now, phase.TimeRangeConfig) || !inDailySubmissionWindow(now, phase) {
		return 0, false
	}

	start, err := parseBizTime(phase.Start)
	if err != nil {
		return 0, false
	}
	day := daysBetween(start, now)
	end, err := parseBizTime(phase.End)
	return day, err == nil && day >= 0 && day <= daysBetween(start, end)
}

func DailyTeamLimit(routeName string, day int) (int, bool) {
	limits, ok := BizConf.DailyTeamLimits[routeName]
	if !ok || day < 0 || day >= len(limits) {
		return 0, false
	}
	return limits[day], true
}

func phaseTimeRange(phase BizPhase) (TimeRangeConfig, bool) {
	switch phase {
	case PhaseRegistration:
		return BizConf.Phases.Registration, true
	case PhaseSubmission:
		return BizConf.Phases.Submission.TimeRangeConfig, true
	case PhaseAdjustment:
		return BizConf.Phases.Adjustment, true
	case PhasePreparation:
		return BizConf.Phases.Preparation, true
	case PhaseActivity:
		return BizConf.Phases.Activity, true
	default:
		return TimeRangeConfig{}, false
	}
}

func inDailySubmissionWindow(now time.Time, phase SubmissionPhaseConfig) bool {
	start, err := time.Parse("15:04:05", phase.DailyStartTime)
	if err != nil {
		return false
	}
	end, err := time.Parse("15:04:05", phase.DailyEndTime)
	if err != nil {
		return false
	}
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), start.Hour(), start.Minute(), start.Second(), 0, now.Location())
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), end.Hour(), end.Minute(), end.Second(), 0, now.Location())
	return !now.Before(todayStart) && !now.After(todayEnd)
}

func inTimeRange(now time.Time, period TimeRangeConfig) bool {
	start, err := parseBizTime(period.Start)
	if err != nil {
		return false
	}
	end, err := parseBizTime(period.End)
	return err == nil && !now.Before(start) && !now.After(end)
}

func parseBizTime(value string) (time.Time, error) {
	return time.ParseInLocation(time.DateTime, value, time.Local)
}

func daysBetween(start, end time.Time) int {
	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
	return int(endDate.Sub(startDate).Hours() / 24)
}
