package apa

import (
	"fmt"
	"strconv"
	"time"

	"github.com/innerr/tiperf/apa/sources"
)

type AutoPerfAssistant struct {
	data map[string]sources.Source
	con  Console
}

func NewAutoPerfAssistant(verbLevel string) *AutoPerfAssistant {
	return &AutoPerfAssistant{
		make(map[string]sources.Source),
		NewConsole(verbLevel),
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
	periods, err = a.detectHardPeriods()
	if err != nil {
		return
	}

	periods, err = a.removeMeaninglessPeriods(periods)
	if err != nil {
		return
	}

	var softs []Period
	for _, period := range periods {
		step := chooseStep(period.End.Sub(period.Start))
		var soft []Period
		soft, err = a.detectSoftPeriods(period, step)
		if err != nil {
			return
		}
		if len(soft) == 0 {
			softs = append(softs, period)
		} else {
			softs = append(softs, soft...)
		}
	}

	periods = softs
	return
}

func (a *AutoPerfAssistant) detectHardPeriods() (periods []Period, err error) {
	// TODO: duration => arg:minDuration
	duration := time.Hour
	maxDuration := 30 * 24 * time.Hour

	periods = make([]Period, 0)
	hards := getPeriodHardBreakingSource()
	end := time.Now()

	var points []BreakingPoint

	for duration <= maxDuration {
		var vectors []CollectedSourceTasks
		vectors, err = collectSources(a.data, hards, end.Add(-duration), end, 0)
		if err != nil {
			return
		}
		for _, vector := range vectors {
			p := findBreakingPoints(vector)
			if len(p) > 0 {
				points = append(points, p...)
			}
		}
		if len(points) > 0 {
			break
		} else {
			duration = duration * 3 / 2
		}
	}

	start := end.Add(-duration)
	if len(points) == 0 {
		periods = []Period{Period{start, end, newBreakingPoint(start), newBreakingPoint(end)}}
	} else {
		periods = genPeriods(start, end, points, time.Minute)
	}
	return
}

func (a *AutoPerfAssistant) removeMeaninglessPeriods(origin []Period) (periods []Period, err error) {
	if len(origin) > 1 {
		origin = origin[1:]
	}
	periods = origin
	return
}

func (a *AutoPerfAssistant) detectSoftPeriods(period Period, step time.Duration) (periods []Period, err error) {
	if period.End.Sub(period.Start) < time.Minute {
		return
	}

	softs := getPeriodSoftBreakingPointSource()
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
		if similarity < SoftPeriodThreshold {
			// TODO: use struct
			reason := fmt.Sprintf("%s vs %s => workload changed: %v", ms2Time(times[i-1]).Format(timeFormat), ms2Time(times[i]).Format(timeFormat), similarity)
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
		err = f(period)
		if err != nil {
			return
		}
	}
	return
}

func (a *AutoPerfAssistant) DetectAll(period Period) (err error) {
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
	a.con.Debug(period, "\n")
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
