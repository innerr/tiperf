package sources

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Prometheus struct {
	client  v1.API
	metrics []string
}

func NewPrometheus(address string) (p *Prometheus, err error) {
	resp, err := http.Get(address + "/api/v1/label/__name__/values")
	if err != nil {
		return
	}
	type Metrics struct {
		Status string
		Data   []string
	}
	var metrics Metrics
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	if err != nil {
		return
	}
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return
	}
	p = &Prometheus{
		v1.NewAPI(client),
		metrics.Data,
	}
	return
}

func (p *Prometheus) Query(query string, start time.Time, end time.Time, step time.Duration) (model.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rng := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}
	res, _, err := p.client.QueryRange(ctx, query, rng)
	return res, err
}

func (p *Prometheus) PreciseQuery(query string, start time.Time, end time.Time) (val model.Value, err error) {
	step := 15 * time.Second
	for {
		val, err = p.Query(query, start, end, step)
		if err == nil {
			return
		}
		if strings.Index(err.Error(), "exceeded maximum resolution of") < 0 {
			return
		}
		step *= 2
	}
}
