package apa

// The content in this file should be put into config file

type SourceTask struct {
	Source   string
	Query    string
	Function string
}

func getPeriodHardBreakingSource() []SourceTask {
	return []SourceTask{
		SourceTask{
			"prometheus",
			"up",
			"eq",
		},
	}
}

// TODO: improve: same metric but different tags could only fetch once
func getPeriodSoftBreakingPointSource() []SourceTask {
	return []SourceTask{
		SourceTask{
			"prometheus",
			"tikv_grpc_msg_duration_seconds_count{type=\"kv_commit\"}",
			"cosine",
		},
		SourceTask{
			"prometheus",
			"tikv_grpc_msg_duration_seconds_count{type=\"kv_prewrite\"}",
			"cosine",
		},
		SourceTask{
			"prometheus",
			"tikv_grpc_msg_duration_seconds_count{type=\"kv_pessimistic_lock\"}",
			"cosine",
		},
		SourceTask{
			"prometheus",
			"tikv_grpc_msg_duration_seconds_count{type=\"coprocessor\"}",
			"cosine",
		},
		SourceTask{
			"prometheus",
			"tikv_grpc_msg_duration_seconds_count{type=\"kv_batch_get_command\"}",
			"cosine",
		},
		SourceTask{
			"prometheus",
			"tikv_grpc_msg_duration_seconds_count{type=\"kv_batch_get\"}",
			"cosine",
		},
	}
}
