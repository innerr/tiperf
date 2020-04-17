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
	case "normal":
		return Console{verbLevelNormal}
	case "compact":
		return Console{verbLevelCompact}
	}
	panic("unknown verb level: '" + verbLevel + "', should be: debug, normal, compact")
}

func (c Console) Debug(msg ...interface{}) {
	if c.verbLevel > verbLevelDebug {
		return
	}
	fmt.Print(msg...)
}

func (c Console) Normal(msg ...interface{}) {
	if c.verbLevel > verbLevelNormal {
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
	verbLevelNormal  = 1
	verbLevelCompact = 2
)
