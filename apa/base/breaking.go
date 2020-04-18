package base

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

type BreakingPoint struct {
	Point  model.Time
	Prev   model.SamplePair
	Curr   model.SamplePair
	Metric model.Metric
}

func (b BreakingPoint) String() string {
	if b.Prev.Timestamp == 0 && b.Curr.Timestamp == 0 && len(b.Metric) == 0 {
		return fmt.Sprintf("(reached border)")
	} else {
		return fmt.Sprintf("%v => %v %v", b.Prev.Value, b.Curr.Value, b.Metric)
	}
}

func NewBreakingPoint(t time.Time) BreakingPoint {
	return BreakingPoint{
		Time2Ms(t),
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

func FindBreakingPoints(vector CollectedSourceTasks) (points []BreakingPoint) {
	eq := GetBreakingFunc(vector.Source.Function)
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

type BreakingFunc func(a model.SampleValue, b model.SampleValue) bool

func PreciseEq(a model.SampleValue, b model.SampleValue) bool {
	// TODO: float cmp
	return a == b
}

func GetBreakingFunc(name string) BreakingFunc {
	switch name {
	case "eq":
		return PreciseEq
	}
	panic("no function: " + name)
}
