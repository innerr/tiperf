package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/innerr/tiperf/apa"
	"github.com/innerr/tiperf/apa/base"

	"github.com/spf13/cobra"
)

var (
	host string
	port int
	verb string

	from     string
	to       string
	duration time.Duration
	period   int
)

func main() {
	var cmd = &cobra.Command{
		Use:   "tiperf",
		Short: "A tool kit help to find perf issues in TiDB by parsing it's prometheus(and other) metrics",
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	cmd.PersistentFlags().StringVarP(&host, "host", "H", "127.0.0.1", "Prometheus host")
	cmd.PersistentFlags().IntVarP(&port, "port", "P", 9090, "Prometheus port")

	cmd.PersistentFlags().StringVar(&verb, "verb", "debug", "Ouput level, sould be: debug|detail|compact")

	cmd.PersistentFlags().StringVarP(&from, "from", "f", "", "Analyze from this time, format: 2006-01-02 15:04:05")
	cmd.PersistentFlags().StringVarP(&to, "to", "t", "", "Analyze to this time, format: 2006-01-02 15:04:05")
	cmd.PersistentFlags().DurationVarP(&duration, "duration", "d", 0, "Analyze from `duration` ago to now, examples: 1h, 30m")
	cmd.PersistentFlags().IntVarP(&period, "period", "p", 0, "A period is a time span runs alike workload. Analyze the last N period")

	registerTimeline(cmd)

	// TODO: more commands

	cmd.Execute()
}

func newAutoPerfAssistant() *apa.AutoPerfAssistant {
	timeRange, err := base.NewTimeRangeFromArgs(from, to, duration)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	apa := apa.NewAutoPerfAssistant(verb, timeRange, period)
	err = apa.AddPrometheus(host, port)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}
	return apa
}

func callHandleFunc(f func() error) {
	err := f()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(3)
	}
}

func registerTimeline(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use: "timeline",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "balance",
		Short: "detect data is balanced in the cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			callHandleFunc(func() error {
				apa := newAutoPerfAssistant()
				return apa.DoDectect(apa.DetectUnbalanced)
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "trend",
		Short: "detect cluster performance trend",
		Run: func(cmd *cobra.Command, _ []string) {
			callHandleFunc(func() error {
				apa := newAutoPerfAssistant()
				return apa.DoDectect(apa.DetectTrend)
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "spike",
		Short: "detect cluster performance spike",
		Run: func(cmd *cobra.Command, _ []string) {
			callHandleFunc(func() error {
				apa := newAutoPerfAssistant()
				return apa.DoDectect(apa.DetectSpike)
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "detect all info in the cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			callHandleFunc(func() error {
				apa := newAutoPerfAssistant()
				return apa.DoDectect(apa.DetectAll)
			})
		},
	})

	parent.AddCommand(cmd)
}
