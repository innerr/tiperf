package apa

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

type TimeRange struct {
	From time.Time
	To   time.Time
}

func (t TimeRange) Valid() bool {
	return !t.From.IsZero() && !t.To.IsZero()
}

func (t TimeRange) String() string {
	return fmt.Sprintf("[%s => %s]", t.From.Format(timeFormat), t.To.Format(timeFormat))
}

func NewTimeRangeFromArgs(from string, to string, duration time.Duration) (t TimeRange, err error) {
	if len(from) != 0 {
		t.From, err = time.Parse(timeFormatZ, from+" CST")
		if err != nil {
			return
		}
	}
	if len(to) != 0 {
		t.To, err = time.Parse(timeFormatZ, to+" CST")
		if err != nil {
			return
		}
	}
	if t.Valid() {
		if t.From.After(t.To) {
			t.From, t.To = t.To, t.From
		}
		return
	}
	if duration == 0 {
		return
	}
	now := time.Now()
	if t.To.IsZero() {
		if t.From.After(now) {
			err = fmt.Errorf("the time arg `from` after `now`: %s vs %s",
				t.From.Format(timeFormatZ), now.Format(timeFormatZ))
			return
		}
		t.To = t.From.Add(duration)
		if t.To.After(now) {
			t.To = now
		}
	} else {
		if t.To.After(now) {
			t.To = now
		}
		t.From = t.To.Add(-duration)
	}
	return
}

type Period struct {
	Start       time.Time
	End         time.Time
	StartReason interface{}
	EndReason   interface{}
}

func (p Period) String() string {
	return fmt.Sprintf("[%s => %s]\n    started by: %v\n    ended   by: %v",
		p.Start.Format(timeFormat), p.End.Format(timeFormat), p.StartReason, p.EndReason)
}

func ms2Time(ms model.Time) time.Time {
	return time.Unix(0, int64(1e6*ms))
}

func time2ms(t time.Time) model.Time {
	return model.Time(t.UnixNano() / 1e6)
}

const (
	timeFormat  = "2006-01-02 15:04:05"
	timeFormatZ = timeFormat + " MST"
)
