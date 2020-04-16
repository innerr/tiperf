package apa

// The content in this file should be put into config file

type SourceConf struct {
	Source string
	Query  string
}

func getPeriodHardBreakingPointSource() []SourceConf {
	return []SourceConf{
		SourceConf{
			"prometheus",
			"up",
		},
		//SourceConf {
		//	"prometheus",
		//	"tikv_grpc_msg_duration_seconds_count{type=\"kv_commit\"}",
		//},
	}
}
