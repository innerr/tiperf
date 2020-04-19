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
	step = times[1].Sub(times[0])

	points = []time.Time{period.Start}
	reasons = []interface{}{period.StartReason}

	for i := 1; i < len(similarities); i++ {
		similarity := similarities[i]
		if similarity >= similarityThreshold {
			continue
		}

		// TODO: Could be improved here, it may fall into a locally optimal point
		if i+1 < len(similarities) && similarities[i+1] < similarity {
			con.Debug("## similarity falling to next ", similarity, similarities[i+1], "\n")
			continue
		}

		//con.Debug("## before zoom in ", times[i-1].Format(TimeFormat), " => ", times[i].Add(step).Format(TimeFormat),
		//	", step ", step, ", similarity", similarity, "\n")

		preciseTime, zoomedStep, zoomedSimilarity, zoomed, err := ZoomInBySimilarity(data, sources,
			times[i-1].Add(-step), times[i].Add(step*2), zoomInSpeed, step/time.Duration(zoomInSpeed), time.Minute, 0, con)
		if err != nil {
			return nil, nil, err
		}

		if zoomed {
			reasons = append(reasons, SimilarityBreakingReason{
				"workload",
				preciseTime.Add(-zoomedStep),
				preciseTime,
				zoomedStep,
				zoomedSimilarity,
			})
			//con.Debug("## after  zoom in ", preciseTime.Add(-zoomedStep).Format(TimeFormat), " => ",
			//	preciseTime.Add(zoomedStep).Format(TimeFormat), ", step ", zoomedStep, ", similarity", zoomedSimilarity, "\n")
			points = append(points, preciseTime)
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
	return
}

func ZoomInBySimilarity(data sources.Sources, sources []SourceTask, start time.Time,
	end time.Time, speed int, step time.Duration, minStep time.Duration, level int, con Console) (preciseTime time.Time,
	zoomedStep time.Duration, zoomedSimilarity float64, zoomed bool, err error) {

	zoomed = false
	if step < minStep {
		return
	}
	similarities, times, err := CollectSimilarities(data, sources, start, end, step)
	if err != nil {
		return
	}
	if len(similarities) < 3 {
		return
	}
	step = times[1].Sub(times[0])
	//for i := 1; i < len(similarities); i++ {
	//	con.Debug("## level ", level, " #", i, " zooming in ", times[i-1].Format(TimeFormat), " => ",
	//		times[i].Add(step).Format(TimeFormat), ", step ", step, ", similarity", similarities[i], "\n")
	//}

	zoomedSimilarity = similarities[0]
	zoomedIndex := -1
	for i := 1; i < len(similarities); i++ {
		if similarities[i] < zoomedSimilarity {
			zoomedSimilarity = similarities[i]
			zoomedIndex = i
		}
	}
	preciseTime = times[zoomedIndex]
	zoomedStep = times[1].Sub(times[0])
	zoomed = true

	rezoomedTime, rezoomedStep, rezoomedSimilarity, rezoomed, err := ZoomInBySimilarity(data, sources,
		preciseTime.Add(-zoomedStep*2), preciseTime.Add(zoomedStep*2), speed, zoomedStep/time.Duration(speed), minStep, level+1, con)
	if err != nil {
		return
	}
	if !rezoomed {
		return
	}

	preciseTime, zoomedStep, zoomedSimilarity = rezoomedTime, rezoomedStep, rezoomedSimilarity
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
