package apa

import (
	"fmt"
	"math"
	"time"

	"github.com/prometheus/common/model"
)

// TODO: have risk that could be too many data and fail, need to split into many queries
func chooseStep(duration time.Duration) time.Duration {
	if duration >= 30*24*time.Hour {
		return 15 * time.Minute
	}
	if duration >= 7*24*time.Hour {
		return 5 * time.Minute
	}
	return time.Minute
}

// PeriodVec represent a period with a multiply dimension vector
type PeriodVec []float64

func alignVectorsLength(origin []CollectedSourceTasks) (vectors []CollectedSourceTasks) {
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

func rotateToPeriodVecs(vectors []CollectedSourceTasks) (vecs []PeriodVec, times []model.Time) {
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
					fmt.Println("WW", i, t, it.Pairs[i].Timestamp)
				}
			}
			vec = append(vec, float64(it.Pairs[i].Value))
		}
		vecs = append(vecs, vec)
		times = append(times, t)
	}
	return
}

func cosineSimilarity(a PeriodVec, b PeriodVec) float64 {
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

const (
	INT_MAX = int(^uint(0) >> 1)
)
