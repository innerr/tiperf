package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/innerr/tiperf/apa"

	"github.com/spf13/cobra"
)

var (
	host string
	port int
	verb string
)

func main() {
	var cmd = &cobra.Command{
		Use:   "tiperf",
		Short: "A tool kit help to find perf issues in TiDB by parsing it's prometheus(and other) metrics",
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	cmd.PersistentFlags().StringVarP(&host, "host", "H", "127.0.0.1", "Prometheus host")
	cmd.PersistentFlags().IntVarP(&port, "port", "P", 9090, "Prometheus port")
	cmd.PersistentFlags().StringVarP(&verb, "verb", "v", "debug", "Ouput level, sould be: debug|normal|compact")

	registerDetect(cmd)
	registerWatch(cmd)

	cmd.Execute()
}

func newAutoPerfAssistant() *apa.AutoPerfAssistant {
	apa := apa.NewAutoPerfAssistant(verb)
	err := apa.AddPrometheus(host, port)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return apa
}

func callHandleFunc(f func() error) {
	err := f()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func registerDetect(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use: "detect",
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

func registerWatch(rootCmd *cobra.Command) {
	// TODO
}
