package detectors

import (
	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/sources"
)

// TODO: impl
func DetectTrend(data sources.Sources, period base.Period, found FoundEvents, con base.Console) (Events, error) {
	/*
		sources := base.GetPeriodWorkloadBreakingPointSource()
		vectors, err := base.CollectSources(data, sources, period.Start, period.End, 0)
		if err != nil {
			return
		}
	*/
	return nil, nil
}
