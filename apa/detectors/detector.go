package detectors

import (
	"time"

	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/sources"
)

type EventInfo interface {
	Output(when time.Time, con base.Console, indent string)
}

type Event struct {
	When time.Time
	What EventInfo
}

type Events []Event

func (e Events) Len() int {
	return len(e)
}

func (e Events) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e Events) Less(i, j int) bool {
	return e[i].When.Before(e[j].When)
}

type Detector func(sources map[string]sources.Source, period base.Period) (Events, error)
