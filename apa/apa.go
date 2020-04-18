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
	data map[string]sources.Source
	con  base.Console

	timeRange   base.TimeRange
	periodCount int
}

func NewAutoPerfAssistant(verbLevel string, timeRange base.TimeRange, periodCount int) *AutoPerfAssistant {
	return &AutoPerfAssistant{
		make(map[string]sources.Source),
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
		period := base.Period{start, end, "analyze start time", "analyze end time"}
		step := base.ChooseStep(duration)
		periods, err = a.detectWorkloadPeriods(period, step)
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

func (a *AutoPerfAssistant) detectWorkloadPeriods(period base.Period, step time.Duration) (periods []base.Period, err error) {
	if period.End.Sub(period.Start) < time.Minute {
		return
	}

	a.con.Debug("## ", period.Start.Format(base.TimeFormat), " => ",
		period.End.Format(base.TimeFormat), " detecting worload periods\n")

	sources := base.GetPeriodWorkloadBreakingPointSource()
	vectors, err := base.CollectSources(a.data, sources, period.Start, period.End, step)
	if err != nil {
		return
	}
	if len(vectors) == 0 {
		return
	}

	vectors = base.AlignVectorsLength(vectors)
	if len(vectors[0].Pairs) == 0 {
		return
	}
	vecs, times := base.RotateToPeriodVecs(vectors)

	points := []time.Time{period.Start}
	reasons := []interface{}{period.StartReason}
	prevTime := times[0]
	for i := 1; i < len(vecs); i++ {
		similarity := base.CosineSimilarity(vecs[i-1], vecs[i])
		if similarity < base.WorkloadPeriodThreshold {
			if (times[i]-prevTime) > 60*1000 || true {
				// TODO: When step is too big, detect with smaller step
				// TODO: use struct
				reason := fmt.Sprintf("%s vs %s => workload changed, comparing-step %v, similarity %v",
					base.Ms2Time(times[i-1]).Format(base.TimeFormat), base.Ms2Time(times[i]).Format(base.TimeFormat), step, similarity)
				points = append(points, base.Ms2Time(times[i]))
				reasons = append(reasons, reason)
			}
			prevTime = times[i]
		}
	}
	points = append(points, period.End)
	reasons = append(reasons, period.EndReason)

	if len(points) <= 2 {
		return
	}
	for i := 1; i < len(points); i++ {
		periods = append(periods, base.Period{points[i-1], points[i], reasons[i-1], reasons[i]})
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
		a.con.Debug("    ## started by: ", period.StartReason, "\n")
		var events detectors.Events
		events, err = detector.RunWorkload(a.data, period)
		if err != nil {
			return
		}
		for _, event := range events {
			a.con.Debug("    ## ", event.When.Format(base.TimeFormat), " ", event.What, "\n")
		}
		a.con.Debug("    ## ended   by: ", period.EndReason, "\n")
	}
	return
}
