package base

import (
	"fmt"
	"time"

	"github.com/innerr/tiperf/apa/sources"
)

type Period struct {
	Start       time.Time
	End         time.Time
	StartReason interface{}
	EndReason   interface{}
}

func CollectPrecisePointsBySimilarity(data sources.Sources, sources []SourceTask, period Period, step time.Duration,
	similarityThreshold float64, zoomInSpeed int, con Console) (points []time.Time, reasons []interface{}, err error) {

	vectors, err := CollectSources(data, sources, period.Start, period.End, step)
	if err != nil || len(vectors) == 0 {
		return
	}
	similarities, rawVecs, times := CalculateSimilarities(vectors)
	if len(similarities) < 2 {
		return
	}

	// Step may be ajusted
	step = times[1].Sub(times[0])

	rawPoints := []time.Time{}
	rawReasons := []SimilarityBreakingReason{}

	for i := 1; i < len(similarities); i++ {
		similarity := similarities[i]
		if similarity >= similarityThreshold {
			continue
		}

		// Double the range to make sure nothing missed
		zoomInStart := times[i-1].Add(-step)
		zoomInEnd := times[i].Add(step * 2)

		const minStep = 15 * time.Second
		zoomInStep := step / time.Duration(zoomInSpeed)
		if zoomInStep < minStep {
			zoomInStep = minStep
		}

		var zoomedSimilarities []float64
		var zoomedTimes []time.Time
		var zoomedStep time.Duration
		zoomed := false
		if zoomInStep != step {
			con.Debug("## before zoom-in ", zoomInStart.Format(TimeFormat), " => ", zoomInEnd.Format(TimeFormat),
				", parent-step ", step, ", parent-similarity ", similarity, ", zoom-in-step ", zoomInStep, "\n")
			zoomedSimilarities, zoomedTimes, zoomedStep, zoomed, err = ZoomInBySimilarity(data, sources,
				zoomInStart, zoomInEnd, similarityThreshold, zoomInSpeed, zoomInStep, minStep, 0, con)
			if err != nil {
				return nil, nil, err
			}
		}

		if zoomed {
			for i, preciseTime := range zoomedTimes {
				zoomedPrevStart := preciseTime.Add(-zoomedStep)
				rawReasons = append(rawReasons, SimilarityBreakingReason{
					"workload",
					zoomedPrevStart,
					preciseTime,
					zoomedStep,
					zoomedSimilarities[i],
				})
				rawPoints = append(rawPoints, preciseTime)
				con.Debug("## after zoom-in, result #", i, " ", zoomedPrevStart.Format(TimeFormat), " => ",
					preciseTime.Add(zoomedStep).Format(TimeFormat), ", step ", zoomedStep,
					", similarity ", zoomedSimilarities[i], "\n")
			}
		} else {
			rawReasons = append(rawReasons, SimilarityBreakingReason{
				"workload",
				times[i-1],
				times[i],
				step,
				similarity,
			})
			rawPoints = append(rawPoints, times[i])
		}
	}

	// Remove duplicated points and out of range points
	dedPoints := []time.Time{}
	dedReasons := []SimilarityBreakingReason{}
	var prev time.Time
	for i, point := range rawPoints {
		if point.Before(period.Start) || point.After(period.End) {
			con.Debug("## throw away point ", point.Format(TimeFormat), "\n")
			continue
		}
		if prev.IsZero() {
			dedPoints = append(dedPoints, rawPoints[i])
			dedReasons = append(dedReasons, rawReasons[i])
			prev = rawPoints[i]
		} else {
			if rawPoints[i].Sub(prev) > PikeDurationMax {
				dedPoints = append(dedPoints, rawPoints[i])
				dedReasons = append(dedReasons, rawReasons[i])
				prev = rawPoints[i]
			}
		}
	}

	if len(dedPoints) == 0 {
		return
	}

	descs := CaculateWorkloadDescs(rawVecs, dedPoints)
	if len(descs) != len(dedPoints)+1 {
		panic(fmt.Sprintf("len(workload descs) should be len(points)+1, got: %v vs %v", len(descs), len(dedPoints)))
	}

	// TODO: Remove inactive points
	points = []time.Time{period.Start}
	reasons = []interface{}{period.StartReason}
	for i, point := range dedPoints {
		points = append(points, point)
		reasons = append(reasons, WorkloadBreakingReason{dedReasons[i], descs[i], descs[i+1]})
	}
	points = append(points, period.End)
	reasons = append(reasons, period.EndReason)

	return
}

func ZoomInBySimilarity(data sources.Sources, sources []SourceTask, start time.Time, end time.Time, similarityThreshold float64,
	speed int, step time.Duration, minStep time.Duration, level int, con Console) (zoomedSimilarities []float64,
	zoomedTimes []time.Time, zoomedStep time.Duration, zoomed bool, err error) {

	zoomed = false
	if step < minStep {
		return
	}

	vectors, err := CollectSources(data, sources, start, end, step)
	if err != nil || len(vectors) == 0 {
		return
	}
	rawSimilarities, _, rawTimes := CalculateSimilarities(vectors)

	// Scale too little, not a succeeded zooming
	if len(rawTimes) < 2 {
		return
	}

	similarities := []float64{}
	times := []time.Time{}
	for i, time := range rawTimes {
		if rawSimilarities[i] >= similarityThreshold {
			continue
		}
		con.Debug("## level ", level, " #", i, " zooming-in ", time.Add(-step).Format(TimeFormat), " => ",
			time.Add(step).Format(TimeFormat), ", step ", step, ", similarity ", rawSimilarities[i], "\n")
		if time.Before(start) || time.After(end) {
			con.Debug("## throw away out of range point ", time.Format(TimeFormat), "\n")
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
		rezoomStep := step / time.Duration(speed)
		rezoomedSimilarities, rezoomedTimes, _, rezoomed, rezoomErr := ZoomInBySimilarity(
			data, sources, it.Add(-2*step), it.Add(2*step), similarityThreshold, speed,
			rezoomStep, minStep, level+1, con)
		if rezoomErr != nil {
			err = rezoomErr
			return
		}
		if rezoomed {
			for j, similarity := range rezoomedSimilarities {
				zoomedSimilarities = append(zoomedSimilarities, similarity)
				zoomedTimes = append(zoomedTimes, rezoomedTimes[j])
			}
		} else {
			zoomedSimilarities = append(zoomedSimilarities, similarities[i])
			zoomedTimes = append(zoomedTimes, it)
		}
	}

	// TODO: the zoomedStep maybe wrong, some points may from re-zoom result
	zoomedStep = step
	zoomed = true
	return
}

func CaculateWorkloadDescs(vecs []CollectedSourceTasks, splittingPoints []time.Time) (descs []WorkloadDesc) {
	if len(vecs) == 0 || len(vecs[0].Pairs) == 0 || len(splittingPoints) == 0 {
		return
	}
	vecCount := len(vecs[0].Pairs)

	names := make([]string, len(vecs))
	for i, vec := range vecs {
		names[i] = string(vec.Metric["type"])
	}

	prevIdx := 0
	pointIdx := 0
	vecIdx := 0
	sums := make([]float64, len(vecs))
	for ; vecIdx < vecCount && pointIdx < len(splittingPoints); vecIdx++ {
		time := Ms2Time(vecs[0].Pairs[vecIdx].Timestamp)
		if !time.Before(splittingPoints[pointIdx]) {
			descs = append(descs, NewWorkloadDesc(sums, names, vecIdx-prevIdx))
			sums = make([]float64, len(vecs))
			pointIdx += 1
			prevIdx = vecIdx
		}
		for i, _ := range sums {
			sums[i] += float64(vecs[i].Pairs[vecIdx].Value)
		}
	}

	if pointIdx < len(splittingPoints) {
		panic("some splitting points out of range")
	}

	sums = make([]float64, len(vecs))
	for ; vecIdx < vecCount; vecIdx++ {
		for i, _ := range sums {
			sums[i] += float64(vecs[i].Pairs[vecIdx].Value)
		}
	}
	descs = append(descs, NewWorkloadDesc(sums, names, vecCount-1-prevIdx))
	return
}

type WorkloadDesc struct {
	AvgQpsCoprocessor     float64
	AvgQpsBatchGet        float64
	AvgQpsBatchGetCommand float64
	AvgQpsCommit          float64
	AvgQpsPessimisticLock float64
	AvgQpsPrewrite        float64
}

func (w WorkloadDesc) RawString() string {
	return fmt.Sprintf("%v %v %v %v %v %v",
		w.AvgQpsCoprocessor, w.AvgQpsBatchGet, w.AvgQpsBatchGetCommand,
		w.AvgQpsCommit, w.AvgQpsPessimisticLock, w.AvgQpsPrewrite)
}

func (w WorkloadDesc) String() string {
	read := w.AvgQpsCoprocessor + w.AvgQpsBatchGet + w.AvgQpsBatchGetCommand
	write := w.AvgQpsCommit + w.AvgQpsPessimisticLock + w.AvgQpsPrewrite
	read /= 2
	write /= 2
	total := read + write

	level := "inactive"
	if total >= QpsThresholdHeavy {
		level = "heavy"
	} else if total >= QpsThresholdAlot {
		level = "lots of"
	} else if total >= QpsThresholdActive {
		level = "slight"
	} else {
		return level
	}

	tp := "read and write"
	if write == 0 || read/write > 20 {
		tp = "read"
	} else if read == 0 || write/read > 20 {
		tp = "write"
		if w.AvgQpsPessimisticLock > QpsThresholdActive && w.AvgQpsPessimisticLock/(w.AvgQpsCommit+w.AvgQpsPrewrite) > 0.1 {
			tp = "pessimistic write"
		}
	}

	return level + " " + tp
}

func NewWorkloadDesc(sums []float64, names []string, samples int) (desc WorkloadDesc) {
	for i, name := range names {
		qps := sums[i] / float64(samples)
		switch name {
		case "coprocessor":
			desc.AvgQpsCoprocessor = qps
		case "kv_batch_get":
			desc.AvgQpsBatchGet = qps
		case "kv_batch_get_command":
			desc.AvgQpsBatchGetCommand = qps
		case "kv_commit":
			desc.AvgQpsCommit = qps
		case "kv_pessimistic_lock":
			desc.AvgQpsPessimisticLock = qps
		case "kv_prewrite":
			desc.AvgQpsPrewrite = qps
		}
	}
	return desc
}

type WorkloadBreakingReason struct {
	Similarity   SimilarityBreakingReason
	PrevWorkload WorkloadDesc
	CurrWorkload WorkloadDesc
}

func (w WorkloadBreakingReason) String() string {
	return fmt.Sprintf("from %v to %v (sim: %.2f)", w.PrevWorkload, w.CurrWorkload, w.Similarity.Similarity)
}
