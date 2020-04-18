package apa

import (
	"fmt"
	"sort"
	"time"

	"github.com/prometheus/common/model"

	"github.com/innerr/tiperf/apa/sources"
)

func getVectors(source sources.Source, query string, start time.Time, end time.Time, step time.Duration) (vectors model.Matrix, err error) {
	var res model.Value
	if step == 0 {
		res, err = source.PreciseQuery(query, start, end)
	} else {
		if step < time.Minute {
			step = time.Minute
		}
		res, err = source.Query(query, start, end, step)
	}
	if err != nil {
		return
	}
	vectors = res.(model.Matrix)
	return
}

type CollectedSourceTasks struct {
	Pairs  []model.SamplePair
	Metric model.Metric
	Source SourceTask
}

func collectSources(
	data map[string]sources.Source,
	sources []SourceTask,
	start time.Time,
	end time.Time,
	step time.Duration) (vectors []CollectedSourceTasks, err error) {

	vectors = make([]CollectedSourceTasks, 0)

	for _, it := range sources {
		source, ok := data[it.Source]
		if !ok {
			err = fmt.Errorf("no data source: " + it.Source)
			return
		}
		var matrix model.Matrix
		matrix, err = getVectors(source, it.Query, start, end, step)
		if err != nil {
			return
		}
		for _, v := range matrix {
			vectors = append(vectors, CollectedSourceTasks{v.Values, v.Metric, it})
		}
	}
	return
}

type BreakingPoint struct {
	Point  model.Time
	Prev   model.SamplePair
	Curr   model.SamplePair
	Metric model.Metric
}

func newBreakingPoint(t time.Time) BreakingPoint {
	return BreakingPoint{
		time2ms(t),
		model.SamplePair{},
		model.SamplePair{},
		model.Metric{},
	}
}

type BreakingPoints []BreakingPoint

func (b BreakingPoints) Len() int {
	return len(b)
}

func (b BreakingPoints) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b BreakingPoints) Less(i, j int) bool {
	return b[i].Point < b[j].Point
}

func findBreakingPoints(vector CollectedSourceTasks) (points []BreakingPoint) {
	eq := getBreakingFunc(vector.Source.Function)
	var prev model.SamplePair
	for i, pair := range vector.Pairs {
		if i == 0 {
			prev = pair
			continue
		}
		sameValue := eq(pair.Value, prev.Value)
		if sameValue {
			continue
		}
		points = append(points, BreakingPoint{pair.Timestamp, prev, pair, vector.Metric})
		prev = pair
	}
	return
}

func genPeriods(start time.Time, end time.Time, points BreakingPoints, minStep time.Duration) (periods []Period) {
	if len(points) <= 0 {
		return
	}

	startPoint := newBreakingPoint(start)
	endPoint := newBreakingPoint(end)
	points = append(points, endPoint)

	sort.Sort(points)
	prev := startPoint
	for _, point := range points {
		prevTime := ms2Time(prev.Point)
		currTime := ms2Time(point.Point)
		if currTime.Sub(prevTime) <= minStep {
			continue
		}
		periods = append(periods, Period{
			prevTime,
			currTime,
			PeriodEndReasonHard{
				prev.Prev,
				prev.Curr,
				prev.Metric,
			},
			PeriodEndReasonHard{
				point.Prev,
				point.Curr,
				point.Metric,
			},
		})
		prev = point
	}
	return periods
}

type PeriodEndReasonHard struct {
	Prev   model.SamplePair
	Curr   model.SamplePair
	Metric model.Metric
}

func (p PeriodEndReasonHard) String() string {
	if p.Prev.Timestamp == 0 && p.Curr.Timestamp == 0 && len(p.Metric) == 0 {
		return fmt.Sprintf("(reached border)")
	} else {
		return fmt.Sprintf("%v => %v %v", p.Prev.Value, p.Curr.Value, p.Metric)
	}
}
