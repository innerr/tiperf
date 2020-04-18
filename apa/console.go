package apa

import (
	"fmt"
)

type Console struct {
	verbLevel int
}

func NewConsole(verbLevel string) Console {
	switch verbLevel {
	case "debug":
		return Console{verbLevelDebug}
	case "detail":
		return Console{verbLevelDetail}
	case "compact":
		return Console{verbLevelCompact}
	}
	panic("unknown verb level: '" + verbLevel + "', should be: debug, detail, compact")
}

func (c Console) Debug(msg ...interface{}) {
	if c.verbLevel > verbLevelDebug {
		return
	}
	fmt.Print(msg...)
}

func (c Console) Detail(msg ...interface{}) {
	if c.verbLevel > verbLevelDetail {
		return
	}
	fmt.Print(msg...)
}

func (c Console) Compact(msg ...interface{}) {
	if c.verbLevel > verbLevelCompact {
		return
	}
	fmt.Print(msg...)
}

const (
	verbLevelDebug   = 0
	verbLevelDetail  = 1
	verbLevelCompact = 2
)
