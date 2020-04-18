package base

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"

	"github.com/innerr/tiperf/apa/sources"
)

func GetVectors(source sources.Source, query string,
	start time.Time, end time.Time, step time.Duration) (vectors model.Matrix, err error) {

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

func CollectSources(
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
		matrix, err = GetVectors(source, it.Query, start, end, step)
		if err != nil {
			return
		}
		for _, v := range matrix {
			vectors = append(vectors, CollectedSourceTasks{v.Values, v.Metric, it})
		}
	}
	return
}

// TODO: have risk that could be too many data and fail, need to split into many queries
func ChooseStep(duration time.Duration) time.Duration {
	if duration >= 30*24*time.Hour {
		return 15 * time.Minute
	}
	if duration >= 7*24*time.Hour {
		return 5 * time.Minute
	}
	return time.Minute
}
