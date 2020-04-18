package base

import (
	"time"
)

type Period struct {
	Start       time.Time
	End         time.Time
	StartReason interface{}
	EndReason   interface{}
}
