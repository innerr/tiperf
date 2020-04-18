package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/innerr/tiperf/apa"
	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/detectors"

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
	runtime.GOMAXPROCS(runtime.NumCPU())

	var cmd = &cobra.Command{
		Use:   "tiperf",
		Short: "Analyze performance of TiDB by parsing it's prometheus(and other) metrics",
	}

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
		Use:   "timeline",
		Short: "Analyze cluster and report in timeline",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			dectectors := detectors.NewDetectors()
			if len(args) == 0 {
				fmt.Println("Usage: append 'name' to select features, '~name' to filter features")
				fmt.Println("Feature list:")
				help := dectectors.HelpStrings()
				for _, h := range help {
					fmt.Println("  " + h)
				}
				return
			}
			apa := newAutoPerfAssistant()
			callHandleFunc(func() (err error) {
				err = dectectors.ParseWorkloadFromArgs(args)
				if err != nil {
					return
				}
				return apa.DoDectect(dectectors)
			})
		},
	}
	parent.AddCommand(cmd)
}
