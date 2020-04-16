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

func (a *AutoPerfAssistant) collectSources(sources []SourceConf, start time.Time, end time.Time) (vectors [][]model.SamplePair, err error) {
	vectors = make([][]model.SamplePair, 0)
	for _, it := range sources {
		source, ok := a.data[it.Source]
		if !ok {
			err = fmt.Errorf("no data source: " + it.Source)
			return
		}
		var v [][]model.SamplePair
		v, err = getVectors(source, it.Query, start, end)
		if err != nil {
			return
		}
		vectors = append(vectors, v...)
	}
	return
}

func (a *AutoPerfAssistant) DetectPeriods() (periods []Period, err error) {
	periods = make([]Period, 0)

	hards := getPeriodHardBreakingPointSource()
	duration := time.Hour
	var points []model.Time

	for duration <= 30*24*time.Hour {
		var vectors [][]model.SamplePair
		vectors, err = a.collectSources(hards, time.Now().Add(-duration), time.Now())
		if err != nil {
			return
		}
		for _, vector := range vectors {
			p := findBreakingPoints(vector, preciseEq)
			if p != nil && len(p) > 0 {
				points = append(points, p...)
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
		periods = genPeriods(time.Now().Add(-duration), points, time.Minute)
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
