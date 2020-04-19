package base

import (
	"math"
	"time"

	"github.com/prometheus/common/model"

	"github.com/innerr/tiperf/apa/sources"
)

func CollectPrecisePointsBySimilarity(data sources.Sources, sources []SourceTask, period Period, step time.Duration,
	similarityThreshold float64, zoomInSpeed int, con Console) (points []time.Time, reasons []interface{}, err error) {

	similarities, times, err := CollectSimilarities(data, sources, period.Start, period.End, step)
	if err != nil || len(similarities) < 2 {
		return
	}

	// Step may be ajusted
	step = times[1].Sub(times[0])

	points = []time.Time{period.Start}
	reasons = []interface{}{period.StartReason}

	for i := 1; i < len(similarities); i++ {
		similarity := similarities[i]
		if similarity >= similarityThreshold {
			continue
		}

		// Double the range to make sure nothing missed
		zoomInStart := times[i-1].Add(-step)
		zoomInEnd := times[i].Add(step * 2)

		const minStep = time.Minute
		zoomInStep := step / time.Duration(zoomInSpeed)
		if zoomInStep < minStep {
			zoomInStep = minStep
		}

		var zoomedSimilarities []float64
		var zoomedTimes []time.Time
		var zoomedStep time.Duration
		var zoomed bool
		if zoomInStep != step {
			con.Debug("## before zoom-in ", zoomInStart.Format(TimeFormat), " => ", zoomInEnd.Format(TimeFormat),
				", parent-step ", step, ", parent-similarity ", similarity, "\n")
			zoomedSimilarities, zoomedTimes, zoomedStep, zoomed, err = ZoomInBySimilarity(data, sources,
				zoomInStart, zoomInEnd, similarityThreshold, zoomInSpeed, zoomInStep, time.Minute, 0, con)
			if err != nil {
				return nil, nil, err
			}
		}

		if zoomed {
			for i, preciseTime := range zoomedTimes {
				zoomedPrevStart := preciseTime.Add(-zoomedStep)
				reasons = append(reasons, SimilarityBreakingReason{
					"workload",
					zoomedPrevStart,
					preciseTime,
					zoomedStep,
					zoomedSimilarities[i],
				})
				points = append(points, preciseTime)
				con.Debug("## after zoom-in ", zoomedPrevStart.Format(TimeFormat), " => ",
					preciseTime.Add(zoomedStep).Format(TimeFormat), ", step ", zoomedStep, ", similarity", zoomedSimilarities[i], "\n")
			}
		} else {
			reasons = append(reasons, SimilarityBreakingReason{
				"workload",
				times[i-1],
				times[i],
				step,
				similarity,
			})
			points = append(points, times[i])
		}
	}

	points = append(points, period.End)
	reasons = append(reasons, period.EndReason)

	// Remove duplicated points and out of range points
	points, rawPoints := []time.Time{}, points
	reasons, rawReasons := []interface{}{}, reasons
	var prev time.Time
	for i, point := range rawPoints {
		if point.Before(period.Start) || point.After(period.End) {
			continue
		}
		if prev.IsZero() {
			points = append(points, rawPoints[i])
			reasons = append(reasons, rawReasons[i])
			prev = rawPoints[i]
		} else {
			if rawPoints[i].Sub(prev) > PikeDurationMax {
				points = append(points, rawPoints[i])
				reasons = append(reasons, rawReasons[i])
				prev = rawPoints[i]
			}
		}
	}

	return
}

func ZoomInBySimilarity(data sources.Sources, sources []SourceTask, start time.Time, end time.Time, similarityThreshold float64,
	speed int, step time.Duration, minStep time.Duration, level int, con Console) (zoomedSimilarities []float64,
	zoomedTimes []time.Time, zoomedStep time.Duration, zoomed bool, err error) {

	zoomed = false
	if step < minStep {
		return
	}
	rawSimilarities, rawTimes, err := CollectSimilarities(data, sources, start, end, step)
	if err != nil {
		return
	}

	// Scale too little, not a succeeded zooming
	if len(rawTimes) < 3 {
		return
	}

	// Step may be ajusted
	step = rawTimes[1].Sub(rawTimes[0])

	similarities := []float64{}
	times := []time.Time{}
	for i, time := range rawTimes {
		con.Debug("## level ", level, " #", i, " zooming-in ", time.Add(-step).Format(TimeFormat), " => ",
			time.Add(step).Format(TimeFormat), ", step ", step, ", similarity ", rawSimilarities[i], "\n")
		if time.Before(start) || time.After(end) {
			continue
		}
		if rawSimilarities[i] >= similarityThreshold {
			continue
		}
		similarities = append(similarities, rawSimilarities[i])
		times = append(times, time)
	}

	// Found nothing
	if len(times) == 0 {
		return
	}

	// Looping zoom-in for more precise points
	for i, it := range times {
		rezoomedSimilarities, rezoomedTimes, rezoomedStep, rezoomed, rezoomErr := ZoomInBySimilarity(
			data, sources, it.Add(-2*zoomedStep), it.Add(2*zoomedStep), similarityThreshold, speed,
			step/time.Duration(speed), minStep, level+1, con)
		if rezoomErr != nil {
			err = rezoomErr
			return
		}
		if rezoomed {
			for j, similarity := range rezoomedSimilarities {
				zoomedSimilarities = append(zoomedSimilarities, similarity)
				zoomedTimes = append(zoomedTimes, rezoomedTimes[j])
			}
			// TODO: this is not all right, not all points are succeedly re-zoom-in
			zoomedStep = rezoomedStep
		} else {
			zoomedSimilarities = append(zoomedSimilarities, similarities[i])
			zoomedTimes = append(zoomedTimes, it)
		}
	}

	zoomed = true
	return
}

func CollectSimilarities(data sources.Sources, sources []SourceTask, start time.Time,
	end time.Time, step time.Duration) (similarities []float64, times []time.Time, err error) {

	vectors, err := CollectSources(data, sources, start, end, step)
	if err != nil || len(vectors) == 0 {
		return
	}
	similarities, times = CalculateSimilarities(vectors)
	return
}

func CalculateSimilarities(vectors []CollectedSourceTasks) (similarities []float64, times []time.Time) {
	vectors = AlignVectorsLength(vectors)
	if len(vectors[0].Pairs) == 0 {
		return
	}
	vecs, timestamps := RotateToPeriodVecs(vectors)
	similarities = []float64{1}
	for i := 1; i < len(vecs); i++ {
		similarity := CosineSimilarity(vecs[i-1], vecs[i])
		similarities = append(similarities, similarity)
	}
	times = make([]time.Time, len(timestamps))
	for i, it := range timestamps {
		times[i] = Ms2Time(it)
	}
	return
}

// PeriodVec represent a period with a multiply dimension vector
type PeriodVec []float64

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
					panic("timestamps not matched in multiply vectors from one query")
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

const (
	INT_MAX = int(^uint(0) >> 1)
)
