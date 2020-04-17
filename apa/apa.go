package apa

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/common/model"

	"github.com/innerr/tiperf/apa/sources"
)

type AutoPerfAssistant struct {
	data map[string]sources.Source
}

func NewAutoPerfAssistant() *AutoPerfAssistant {
	return &AutoPerfAssistant{
		make(map[string]sources.Source),
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
	// TODO: getting data could be merge, get a lot data in one time
	var softs []Period
	for _, period := range periods {
		var soft []Period
		soft, err = a.detectSoftPeriods(period)
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

	var points []model.Time

	for duration <= maxDuration {
		var vectors []CollectedSourceTasks
		vectors, err = collectSources(a.data, hards, time.Now().Add(-duration), time.Now())
		if err != nil {
			return
		}
		for _, vector := range vectors {
			p, r := findBreakingPoints(vector.Pairs, getBreakingFunc(vector.Source.Function))
			if len(p) > 0 {
				points = append(points, p...)
				// TODO: track the reasons
				_ = r
				//for _, x := range r {
				//	t := Period{ms2Time(int64(x.Prev.Timestamp)), ms2Time(int64(x.Next.Timestamp))}
				//	fmt.Printf("%v\n    %s\n    %v => %v\n", vector.Metric, t.String(), x.Prev.Value, x.Next.Value)
				//}
			}
		}
		if len(points) > 0 {
			break
		} else {
			duration *= 2
		}
	}

	if len(points) == 0 {
		periods = []Period{
			Period{
				time.Now().Add(-duration),
				time.Now(),
			},
		}
	} else {
		points = append(points, model.Time(time.Now().UnixNano()/1e6))
		periods = genPeriods(time.Now().Add(-duration), points, time.Minute)
	}
	return
}

func (a *AutoPerfAssistant) detectSoftPeriods(period Period) (periods []Period, err error) {
	//softs := getPeriodSoftBreakingPointSource()
	//vectors, err = collectSources(a.data, softs, period.Start, period.End)
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
	fmt.Println(period)
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
