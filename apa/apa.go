package apa

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/innerr/tiperf/apa/sources"
)

type AutoPerfAssistant struct {
	data map[string]sources.Source
	con  Console

	timeRange   TimeRange
	periodCount int
}

func NewAutoPerfAssistant(verbLevel string, timeRange TimeRange, periodCount int) *AutoPerfAssistant {
	return &AutoPerfAssistant{
		make(map[string]sources.Source),
		NewConsole(verbLevel),
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

func (a *AutoPerfAssistant) DetectPeriods() (periods []Period, err error) {
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
		start = end.Add(-AutoModeMaxDuration)
	}

	duration := end.Sub(start)
	step := chooseStep(duration)
	period := Period{start, end, "analyze start time", "analyze end time"}

	for {
		periods, err = a.detectWorkloadPeriods(period, step)
		if err != nil {
			return
		}
		if len(periods) != 0 {
			break
		}
		// Increase duration no matter auto or not
		if duration > AutoModeMaxDuration {
			a.con.Debug("## detected nothing in max duration\n")
			break
		} else {
			a.con.Debug("## detected nothing, increasing duration: ", duration, " => ")
			duration = duration * 3 / 2
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

func (a *AutoPerfAssistant) DetectAlive(period Period) (err error) {
	sources := getPeriodAliveSource()
	vectors, err := collectSources(a.data, sources, period.Start, period.End, 0)
	if err != nil {
		return
	}
	var points BreakingPoints
	for _, vector := range vectors {
		p := findBreakingPoints(vector)
		if len(p) > 0 {
			points = append(points, p...)
		}
	}
	sort.Sort(points)
	for _, point := range points {
		a.con.Debug("    ## service event: ", point, "\n")
	}
	return
}

func (a *AutoPerfAssistant) removePeriods(origin []Period) (periods []Period, err error) {
	if a.periodCount > 0 && len(origin) > a.periodCount {
		a.con.Debug("## too many periods, removing: ", len(origin), " => ")
		origin = origin[len(origin)-a.periodCount:]
		a.con.Debug(len(origin), "\n")
	}
	if a.periodCount == 0 && !a.timeRange.Valid() {
		a.con.Debug("## removing the first period, it's imcompleted\n")
		origin = origin[1:]
	}
	periods = origin
	return
}

func (a *AutoPerfAssistant) detectWorkloadPeriods(period Period, step time.Duration) (periods []Period, err error) {
	if period.End.Sub(period.Start) < time.Minute {
		return
	}

	softs := getPeriodWorkloadBreakingPointSource()
	vectors, err := collectSources(a.data, softs, period.Start, period.End, step)
	if err != nil {
		return
	}

	vectors = alignVectorsLength(vectors)
	vecs, times := rotateToPeriodVecs(vectors)

	points := []time.Time{period.Start}
	reasons := []interface{}{period.StartReason}
	for i := 1; i < len(vecs); i++ {
		similarity := cosineSimilarity(vecs[i-1], vecs[i])
		if similarity < WorkloadPeriodThreshold {
			// TODO: use struct
			reason := fmt.Sprintf("%s vs %s => workload changed: %v",
				ms2Time(times[i-1]).Format(timeFormat), ms2Time(times[i]).Format(timeFormat), similarity)
			points = append(points, ms2Time(times[i]))
			reasons = append(reasons, reason)
		}
	}
	points = append(points, period.End)
	reasons = append(reasons, period.EndReason)

	if len(points) <= 2 {
		return
	}
	for i := 1; i < len(points); i++ {
		periods = append(periods, Period{points[i-1], points[i], reasons[i-1], reasons[i]})
	}
	return
}

func (a *AutoPerfAssistant) DoDectect(f func(period Period) error) (err error) {
	periods, err := a.DetectPeriods()
	if err != nil {
		return
	}
	for _, period := range periods {
		a.con.Detail(period, "\n")
		err = f(period)
		if err != nil {
			return
		}
	}
	return
}

func (a *AutoPerfAssistant) DetectAll(period Period) (err error) {
	err = a.DetectAlive(period)
	if err != nil {
		return
	}
	err = a.DetectUnbalanced(period)
	if err != nil {
		return
	}
	err = a.DetectUnbalanced(period)
	if err != nil {
		return
	}
	err = a.DetectUnbalanced(period)
	if err != nil {
		return
	}
	return
}

func (a *AutoPerfAssistant) DetectUnbalanced(period Period) (err error) {
	return
}

func (a *AutoPerfAssistant) DetectTrend(period Period) (err error) {
	return
}

func (a *AutoPerfAssistant) DetectSpike(period Period) (err error) {
	return
}
