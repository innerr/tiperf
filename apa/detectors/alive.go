package detectors

import (
	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/sources"
)

func DetectAlive(data map[string]sources.Source, period base.Period) (events Events, err error) {
	sources := base.GetPeriodAliveSource()
	vectors, err := base.CollectSources(data, sources, period.Start, period.End, 0)
	if err != nil {
		return
	}
	for _, vector := range vectors {
		points := base.FindBreakingPoints(vector)
		for _, point := range points {
			events = append(events, Event{base.Ms2Time(point.Point), point})
		}
	}
	return
}
