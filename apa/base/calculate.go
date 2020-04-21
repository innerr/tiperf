package base

import (
	"fmt"
	"math"
	"time"

	"github.com/prometheus/common/model"
)

func CalculateSimilarities(vectors []CollectedSourceTasks) (similarities []float64, aligned []CollectedSourceTasks, times []time.Time) {
	vectors = AlignVectorsLength(vectors)
	if len(vectors[0].Pairs) == 0 {
		return
	}
	vecs, timestamps := RotateToPeriodVecs(vectors)
	times = make([]time.Time, len(timestamps))
	for i, it := range timestamps {
		times[i] = Ms2Time(it)
	}
	similarities = []float64{1}
	for i := 1; i < len(vecs); i++ {
		//similarity := DistanceSimilarity(vecs[i-1], vecs[i])
		similarity := CosineSimilarity(vecs[i-1], vecs[i])
		if math.IsNaN(similarity) {
			threshold := QpsThresholdActive * 6 / times[i].Sub(times[i-1]).Seconds()
			z1 := vecs[i-1].Sum() < threshold
			z2 := vecs[i].Sum() < threshold
			if z1 && z2 {
				similarity = 1
			} else if z1 || z2 {
				similarity = 0.5
			} else {
				panic(fmt.Sprintf("similarity is NaN: %v vs %v", vecs[i-1], vecs[i]))
			}
		}
		similarities = append(similarities, similarity)
	}
	aligned = vectors
	return
}

// PeriodVec represent a period's property with a multiply dimension vector
type PeriodVec []float64

func (p PeriodVec) Sum() float64 {
	sum := float64(0)
	for _, it := range p {
		sum += it
	}
	return sum
}

func AlignVectorsLength(origin []CollectedSourceTasks) (vectors []CollectedSourceTasks) {
	if len(origin) == 0 {
		return
	}
	minLen := INT_MAX
	for _, it := range origin {
		if minLen > len(it.Pairs) {
			minLen = len(it.Pairs)
		}
	}

	for _, it := range origin {
		if len(it.Pairs) > minLen {
			it.Pairs = it.Pairs[0:minLen]
			// TODO: need more info merging here?
		}
		vectors = append(vectors, it)
	}
	return vectors
}

func RotateToPeriodVecs(vectors []CollectedSourceTasks) (vecs []PeriodVec, times []model.Time) {
	if len(vectors) == 0 {
		return
	}
	count := len(vectors[0].Pairs)
	for i := 0; i < count; i++ {
		var vec PeriodVec
		var t model.Time
		for j, it := range vectors {
			if j == 0 {
				t = it.Pairs[i].Timestamp
			} else {
				if t != it.Pairs[i].Timestamp {
					line := fmt.Sprintf("%s vs %s", Ms2Time(t).Format(TimeFormat), Ms2Time(it.Pairs[i].Timestamp).Format(TimeFormat))
					panic("timestamps not matched in multiply vectors from one query: " + line)
				}
			}
			vec = append(vec, float64(it.Pairs[i].Value))
		}
		vecs = append(vecs, vec)
		times = append(times, t)
	}
	return
}

func CosineSimilarity(a PeriodVec, b PeriodVec) float64 {
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

func DistanceSimilarity(a PeriodVec, b PeriodVec) float64 {
	var r float64
	r = 0
	for i := 0; i < len(a); i++ {
		r = r + (a[i] * b[i])
	}
	r = math.Sqrt(r)
	return r
}

const (
	INT_MAX = int(^uint(0) >> 1)
)
