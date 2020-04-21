package apa

import (
	"fmt"
	"strconv"
	"time"

	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/detectors"
	"github.com/innerr/tiperf/apa/sources"
)

type AutoPerfAssistant struct {
	data sources.Sources
	con  base.Console

	timeRange   base.TimeRange
	periodCount int
}

func NewAutoPerfAssistant(verbLevel string, timeRange base.TimeRange, periodCount int) *AutoPerfAssistant {
	return &AutoPerfAssistant{
		make(sources.Sources),
		base.NewConsole(verbLevel),
		timeRange,
		periodCount,
	}
}

func (a *AutoPerfAssistant) AddPrometheus(host string, port int) error {
	address := "http://" + host + ":" + strconv.Itoa(port)
	source, err := sources.NewPrometheus(address)
	if err != nil {
		return err
	}
	a.data["prometheus"] = source
	return nil
}

func (a *AutoPerfAssistant) DetectPeriods() (periods []base.Period, err error) {
	autoMode := !a.timeRange.Valid()

	if autoMode {
		a.con.Debug("## args: auto dectect time range to analyze\n")
	} else {
		a.con.Debug("## args: analyze range ", a.timeRange, "\n")
	}
	if a.periodCount != 0 {
		a.con.Debug("## args: analyze last ", a.periodCount, " period(s)\n")
	}

	now := time.Now()

	end := a.timeRange.To
	start := a.timeRange.From
	if autoMode {
		end = now
		start = end.Add(-base.AutoModeStartDuration)
	}
	duration := end.Sub(start)

	for {
		period := base.Period{
			start,
			end,
			"start",
			"end",
		}
		periods, err = detectors.DetectWorkloadPeriods(a.data, period, a.con)
		if err != nil {
			return
		}
		if a.periodCount > 0 && len(periods) > a.periodCount {
			break
		}
		if a.periodCount == 0 && len(periods) != 0 {
			break
		}
		// Increase duration no matter auto or not
		if duration > base.AutoModeMaxDuration {
			if a.periodCount == 0 {
				a.con.Debug("## detected nothing in max duration\n")
			} else {
				a.con.Debug("## detected ", len(periods), "/", a.periodCount, " period(s), reached max duration\n")
			}
			break
		} else {
			if a.periodCount == 0 {
				a.con.Debug("## detected nothing")
			} else {
				a.con.Debug("## detected ", len(periods), "/", a.periodCount, " period(s)")
			}
			a.con.Debug(", increasing duration: ", duration, " => ")
			duration *= 2
			start = end.Add(-duration)
			a.con.Debug(duration, "\n")
		}
	}

	a.con.Debug("## dectected ", len(periods), " periods by workload\n")

	periods, err = a.removePeriods(periods)
	if err != nil {
		return
	}
	return
}

func (a *AutoPerfAssistant) removePeriods(origin []base.Period) (periods []base.Period, err error) {
	if a.periodCount > 0 && len(origin) > a.periodCount {
		a.con.Debug("## too many periods, reducing: ", len(origin), " => ")
		origin = origin[len(origin)-a.periodCount:]
		a.con.Debug(len(origin), "\n")
	} else if !a.timeRange.Valid() {
		if len(origin) > 1 {
			a.con.Debug("## removing the first period, it's imcompleted\n")
			origin = origin[1:]
		}
	} else {
		oldLen := len(origin)
		for len(origin) > 0 {
			if a.timeRange.From.After(origin[0].End) {
				origin = origin[1:]
			} else {
				break
			}
		}
		if oldLen != len(origin) {
			a.con.Debug("## removed ", oldLen-len(origin), " not in analyze range period(s)\n")
		}
	}
	periods = origin
	return
}

func (a *AutoPerfAssistant) DoDectect(detector detectors.Detectors) (err error) {
	periods, err := a.DetectPeriods()
	if err != nil {
		return
	}

	workload := detector.GetWorkload()
	for _, w := range workload {
		a.con.Debug("## args: workload ", w, "\n")
	}

	for _, period := range periods {
		a.con.Detail("[", period.Start.Format(base.TimeFormat), " => ", period.End.Format(base.TimeFormat), "]", "\n")
		whyStart := fmt.Sprintf("%v", period.StartReason)
		if whyStart != "start" {
			whyStartReason := period.StartReason.(base.WorkloadBreakingReason)
			a.con.Debug("    ## ", whyStartReason.Similarity, "\n")
			a.con.Debug("    ## prev workload ", whyStartReason.PrevWorkload.RawString(), "\n")
			a.con.Debug("    ## curr workload ", whyStartReason.CurrWorkload.RawString(), "\n")
			a.con.Detail("    ** started ", whyStartReason.CurrWorkload, " (from ", whyStartReason.PrevWorkload, ")\n")
		}

		var events detectors.Events
		events, err = detector.RunWorkload(a.data, period, a.con)
		if err != nil {
			return
		}
		for _, event := range events {
			event.What.Output(event.When, a.con, "    ")
		}

		whyEnd := fmt.Sprintf("%v", period.EndReason)
		lasted := period.End.Sub(period.Start).Truncate(time.Minute)
		if whyEnd != "end" {
			whyEndReason := period.EndReason.(base.WorkloadBreakingReason)
			a.con.Debug("    ## ", whyEndReason.Similarity, "\n")
			a.con.Debug("    ## curr workload ", whyEndReason.PrevWorkload.RawString(), "\n")
			a.con.Debug("    ## next workload ", whyEndReason.CurrWorkload.RawString(), "\n")
			a.con.Detail("    ** went to ", whyEndReason.CurrWorkload, "\n")
		}
		a.con.Detail("    ** lasted ", lasted, "\n")
	}
	return
}
