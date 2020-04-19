package base

// The content in this file should be put into config file

import (
	"time"
)

const (
	AutoModeMaxDuration     = 30 * 24 * time.Hour
	AutoModeStartDuration   = time.Hour
	WorkloadPeriodThreshold = 0.5
	TimeFormat              = "2006-01-02 15:04:05"
	TimeFormatZ             = TimeFormat + " MST"
)

type SourceTask struct {
	Source   string
	Query    string
	Function string
}

func GetPeriodAliveSource() []SourceTask {
	return []SourceTask{
		SourceTask{
			"prometheus",
			"up",
			"eq",
		},
	}
}

func GetPeriodWorkloadBreakingPointSource() []SourceTask {
	return []SourceTask{
		SourceTask{
			"prometheus",
			"sum(rate(tikv_grpc_msg_duration_seconds_count{type=~\"kv_commit|kv_prewrite|kv_pessimistic_lock|coprocessor|kv_batch_get_command|kv_batch_get\"}[%s])) by (type)",
			"cosine",
		},
	}
}

func ChooseWorkloadPeriodSmoothStep(duration time.Duration) time.Duration {
	if duration >= 30*24*time.Hour {
		return 36 * time.Minute
	}
	if duration >= 20*24*time.Hour {
		return 24 * time.Minute
	}
	if duration >= 10*24*time.Hour {
		return 12 * time.Minute
	}
	if duration >= 5*24*time.Hour {
		return 6 * time.Minute
	}
	return 4 * time.Minute
}
