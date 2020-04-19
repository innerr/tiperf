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
	data sources.Sources,
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
