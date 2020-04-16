package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/innerr/tiperf/apa"

	"github.com/spf13/cobra"
)

func main() {
	var cmd = &cobra.Command{
		Use:   "tiperf",
		Short: "A tool kit help to find perf issues in TiDB by parsing it's prometheus(and other) metrics",
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	var host string
	var port int
	cmd.PersistentFlags().StringVarP(&host, "host", "H", "127.0.0.1", "Prometheus host")
	cmd.PersistentFlags().IntVarP(&port, "port", "P", 9090, "Prometheus port")

	apa := apa.NewAutoPerfAssistant()
	err := apa.AddPrometheus(host, port)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	registerDetect(cmd, apa)
	registerWatch(cmd, apa)

	cmd.Execute()
}

func callHandleFunc(f func() error) {
	err := f()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func registerDetect(parent *cobra.Command, apa *apa.AutoPerfAssistant) {
	cmd := &cobra.Command{
		Use: "detect",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "balance",
		Short: "detect data is balanced in the cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			callHandleFunc(func() error {
				return apa.DoDectect(apa.DetectUnbalanced)
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "trend",
		Short: "detect cluster performance trend",
		Run: func(cmd *cobra.Command, _ []string) {
			callHandleFunc(func() error {
				return apa.DoDectect(apa.DetectTrend)
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "spike",
		Short: "detect cluster performance spike",
		Run: func(cmd *cobra.Command, _ []string) {
			callHandleFunc(func() error {
				return apa.DoDectect(apa.DetectSpike)
			})
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "all",
		Short: "detect all info in the cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			callHandleFunc(func() error {
				return apa.DoDectect(apa.DetectAll)
			})
		},
	})

	parent.AddCommand(cmd)
}

func registerWatch(rootCmd *cobra.Command, apa *apa.AutoPerfAssistant) {
	// TODO
}
