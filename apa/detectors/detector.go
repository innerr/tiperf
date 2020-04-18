package detectors

import (
	"time"

	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/sources"
)

type Event struct {
	When time.Time
	What interface{}
}

type Events []Event

type Detector func(sources map[string]sources.Source, period base.Period) Events
