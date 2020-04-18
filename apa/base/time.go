package base

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
	return fmt.Sprintf("%s => %s", t.From.Format(TimeFormat), t.To.Format(TimeFormat))
}

func NewTimeRangeFromArgs(from string, to string, duration time.Duration) (t TimeRange, err error) {
	if len(from) != 0 {
		t.From, err = time.Parse(TimeFormatZ, from+" CST")
		if err != nil {
			return
		}
	}
	if len(to) != 0 {
		t.To, err = time.Parse(TimeFormatZ, to+" CST")
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
	if !t.Valid() {
		t.To = now
		t.From = t.To.Add(-duration)
	} else if t.To.IsZero() {
		if t.From.After(now) {
			err = fmt.Errorf("the time arg `from` after `now`: %s vs %s",
				t.From.Format(TimeFormatZ), now.Format(TimeFormatZ))
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

func Ms2Time(ms model.Time) time.Time {
	return time.Unix(0, int64(1e6*ms))
}

func Time2Ms(t time.Time) model.Time {
	return model.Time(t.UnixNano() / 1e6)
}
