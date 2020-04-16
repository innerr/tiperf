package apa

import (
	"fmt"
	"sort"
	"time"

	"github.com/prometheus/common/model"

	"github.com/innerr/tiperf/apa/sources"
)

func getVectors(source sources.Source, query string, start time.Time, end time.Time) (vectors [][]model.SamplePair, err error) {
	res, err := source.PreciseQuery(query, start, end)
	if err != nil {
		return
	}
	matrix := res.(model.Matrix)
	vectors = make([][]model.SamplePair, 0)
	for _, v := range matrix {
		vectors = append(vectors, v.Values)
	}
	return
}

func findBreakingPoints(pairs []model.SamplePair, eq func(a model.SampleValue, b model.SampleValue) bool) []model.Time {
	points := make([]model.Time, 0)
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
		prev = pair
	}
	return points
}

func genPeriods(start time.Time, unsorted []model.Time, minStep time.Duration) (periods []Period) {
	var points = make([]int, len(unsorted))
	for i, it := range unsorted {
		points[i] = int(it)
	}
	sort.Ints(points)
	prev := start
	for _, point := range points {
		curr := time.Unix(0, int64(1e6*point))
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

type Period struct {
	Start time.Time
	End   time.Time
}

func (p Period) String() string {
	tf := "2006-01-02 15:04:05"
	return fmt.Sprintf("%s => %s", p.Start.Format(tf), p.End.Format(tf))
}
