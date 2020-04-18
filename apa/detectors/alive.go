package detectors

import (
	"fmt"
	"time"

	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/sources"
)

type AliveInfo struct {
	Instance string
	Type     string
	IsUpping bool
}

func (a AliveInfo) Output(when time.Time, con base.Console, indent string) {
	var action string
	if a.IsUpping {
		action = "up"
	} else {
		action = "down"
	}
	line := fmt.Sprintf("%s%s [%v] -> %s %s", indent, when.Format(base.TimeFormat), a.Type, action, a.Instance)
	con.Detail(line, "\n")
}

func DetectAlive(data map[string]sources.Source, period base.Period) (events Events, err error) {
	sources := base.GetPeriodAliveSource()
	vectors, err := base.CollectSources(data, sources, period.Start, period.End, 0)
	if err != nil {
		return
	}
	for _, vector := range vectors {
		points := base.FindBreakingPoints(vector)
		for _, point := range points {
			info := AliveInfo{
				string(point.Metric["instance"]),
				string(point.Metric["job"]),
				point.Curr.Value == 1,
			}
			events = append(events, Event{base.Ms2Time(point.Point), info})
		}
	}
	return
}
