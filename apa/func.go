package apa

import (
	"github.com/prometheus/common/model"
)

func preciseEq(a model.SampleValue, b model.SampleValue) bool {
	// TODO: float cmp
	return a == b
}
