package sources

import (
	"time"

	"github.com/prometheus/common/model"
)

// We support data source types other than prometheus,
//   but thay need to convert into prometheus-format
type Source interface {

	// A data source may not implemented this method
	Query(query string, start time.Time, end time.Time, step time.Duration) (model.Value, error)

	// A data source must implemented this method
	PreciseQuery(query string, start time.Time, end time.Time) (model.Value, error)
}
