package detectors

import (
	"time"

	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/sources"
)

func DetectWorkloadPeriods(data sources.Sources, period base.Period,
	con base.Console) (periods []base.Period, err error) {

	// Calculating: smoothen -> locate rough positions -> zoom in to get precise points

	duration := period.End.Sub(period.Start)
	if duration < time.Minute {
		return
	}

	con.Debug("## ", period.Start.Format(base.TimeFormat), " => ",
		period.End.Format(base.TimeFormat), " detecting worload periods\n")

	sources := base.GetPeriodWorkloadBreakingPointSource()
	step := base.ChooseWorkloadPeriodSmoothStep(duration)

	points, reasons, err := base.CollectPrecisePointsBySimilarity(data, sources, period, step, base.WorkloadPeriodThreshold, 4, con)
	if err != nil || len(points) <= 2 {
		return
	}

	for i := 1; i < len(points); i++ {
		periods = append(periods, base.Period{points[i-1], points[i], reasons[i-1], reasons[i]})
	}
	return
}
