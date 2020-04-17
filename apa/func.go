package apa

import (
	"github.com/prometheus/common/model"
)

type BreakingFunc func(a model.SampleValue, b model.SampleValue) bool

func preciseEq(a model.SampleValue, b model.SampleValue) bool {
	// TODO: float cmp
	return a == b
}

func getBreakingFunc(name string) BreakingFunc {
	switch name {
	case "eq":
		return preciseEq
	}
	panic("no function: " + name)
}
