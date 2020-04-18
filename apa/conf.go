package apa

// The content in this file should be put into config file

import (
	"time"
)

const (
	AutoModeMaxDuration     = 30 * 24 * time.Hour
	AutoModeStartDuration   = time.Hour
	WorkloadPeriodThreshold = 0.95
)

type SourceTask struct {
	Source   string
	Query    string
	Function string
}

func getPeriodAliveSource() []SourceTask {
	return []SourceTask{
		SourceTask{
			"prometheus",
			"up",
			"eq",
		},
	}
}

func getPeriodWorkloadBreakingPointSource() []SourceTask {
	return []SourceTask{
		SourceTask{
			"prometheus",
			"sum(rate(tikv_grpc_msg_duration_seconds_count{type=~\"kv_commit|kv_prewrite|kv_pessimistic_lock|coprocessor|kv_batch_get_command|kv_batch_get\"}[1m])) by (type)",
			"cosine",
		},
	}
}
