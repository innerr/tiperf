package apa

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/prometheus/common/model"

	"github.com/innerr/tiperf/apa/sources"
)

func getVectors(source sources.Source, query string, start time.Time, end time.Time) (vectors model.Matrix, err error) {
	res, err := source.PreciseQuery(query, start, end)
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
	end time.Time) (vectors []CollectedSourceTasks, err error) {

	vectors = make([]CollectedSourceTasks, 0)

	for _, it := range sources {
		source, ok := data[it.Source]
		if !ok {
			err = fmt.Errorf("no data source: " + it.Source)
			return
		}
		var matrix model.Matrix
		matrix, err = getVectors(source, it.Query, start, end)
		if err != nil {
			return
		}
		for _, v := range matrix {
			vectors = append(vectors, CollectedSourceTasks{v.Values, v.Metric, it})
		}
	}
	return
}

type BreakingReason struct {
	Prev model.SamplePair
	Next model.SamplePair
}

func findBreakingPoints(pairs []model.SamplePair, eq BreakingFunc) (points []model.Time, reasons []BreakingReason) {
	var prev model.SamplePair
	for i, pair := range pairs {
		if i == 0 {
			prev = pair
			continue
		}
		sameValue := eq(pair.Value, prev.Value)
		if sameValue {
			continue
		}
		points = append(points, pair.Timestamp)
		reasons = append(reasons, BreakingReason{prev, pair})
		prev = pair
	}
	return
}

func genPeriods(start time.Time, unsorted []model.Time, minStep time.Duration) (periods []Period) {
	var points = make([]int, len(unsorted))
	for i, it := range unsorted {
		points[i] = int(it)
	}
	sort.Ints(points)
	prev := start
	for _, point := range points {
		curr := ms2Time(int64(point))
		if curr.Sub(prev) <= minStep {
			continue
		}
		periods = append(periods, Period{
			prev,
			curr,
		})
		prev = curr
	}
	return periods
}

func ms2Time(ms int64) time.Time {
	return time.Unix(0, int64(1e6*ms))
}

type Period struct {
	Start time.Time
	End   time.Time
}

func (p Period) String() string {
	tf := "2006-01-02 15:04:05"
	return fmt.Sprintf("%s => %s", p.Start.Format(tf), p.End.Format(tf))
}

func cosineSimilarity(a []float64, b []float64) float64 {
	var (
		aLen  = len(a)
		bLen  = len(b)
		s     = 0.0
		sa    = 0.0
		sb    = 0.0
		count = 0
	)
	if aLen > bLen {
		count = aLen
	} else {
		count = bLen
	}
	for i := 0; i < count; i++ {
		if i >= bLen {
			sa += math.Pow(a[i], 2)
			continue
		}
		if i >= aLen {
			sb += math.Pow(b[i], 2)
			continue
		}
		s += a[i] * b[i]
		sa += math.Pow(a[i], 2)
		sb += math.Pow(b[i], 2)
	}
	return s / (math.Sqrt(sa) * math.Sqrt(sb))
}
