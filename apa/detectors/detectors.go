package detectors

import (
	"fmt"
	"sort"

	"github.com/innerr/tiperf/apa/base"
	"github.com/innerr/tiperf/apa/sources"
)

type Detectors struct {
	names         []string
	helps         []string
	combinedNames []string
	combinedHelps []string
	combineds     map[string][]string
	functions     map[string]Detector
	workload      map[string]Detector
}

func NewDetectors() Detectors {
	d := Detectors{
		make([]string, 0),
		make([]string, 0),
		make([]string, 0),
		make([]string, 0),
		make(map[string][]string),
		make(map[string]Detector),
		make(map[string]Detector),
	}
	d.RegisterAll()
	return d
}

func (d *Detectors) Register(name string, help string, function Detector) {
	d.names = append(d.names, name)
	d.helps = append(d.helps, help)
	d.functions[name] = function
}

func (d *Detectors) RegisterCombined(name string, help string, names []string) {
	d.combinedNames = append(d.combinedNames, name)
	d.combinedHelps = append(d.combinedHelps, help)
	d.combineds[name] = names
}

func (d *Detectors) RegisterAll() {
	d.Register("balance", "detect anything imbalance", DetectBalance)
	d.Register("trend", "detect performance trend", DetectTrend)
	d.Register("pikes", "detect performance pikes", DetectPikes)
	d.Register("alive", "detect service up and down events", DetectAlive)

	d.RegisterCombined("all", "detect all", []string{
		"balance",
		"trend",
		"pikes",
		"alive",
	})
}

func (d *Detectors) HelpStrings() []string {
	maxNameLen := 0
	for _, name := range d.names {
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}
	for _, name := range d.combinedNames {
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}

	helps := make([]string, 0)
	for i, name := range d.names {
		helps = append(helps, padding(name, maxNameLen)+"  "+d.helps[i])
	}
	for i, name := range d.combinedNames {
		helps = append(helps, padding(name, maxNameLen)+"  "+d.combinedHelps[i]+
			fmt.Sprintf(", include: %v", d.combineds[name]))
	}
	return helps
}

func (d *Detectors) ParseWorkloadFromArgs(args []string) error {
	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}

		if arg[0] == '~' {
			name := arg[1:]
			delete(d.workload, name)
			continue
		}

		name := arg
		function, ok := d.functions[name]
		if ok {
			d.workload[name] = function
			continue
		}
		names, ok := d.combineds[name]
		if ok {
			for _, it := range names {
				function, ok := d.functions[it]
				if !ok {
					return fmt.Errorf("unknown name: " + it + " in combination: " + name)
				}
				d.workload[it] = function
			}
			continue
		}
		return fmt.Errorf("unknown name: " + name)
	}
	return nil
}

func (d *Detectors) GetWorkload() (workload []string) {
	for k, _ := range d.workload {
		workload = append(workload, k)
	}
	return
}

// TODO: keep the register order
func (d *Detectors) RunWorkload(sources map[string]sources.Source, period base.Period) (events Events, err error) {
	var ev Events
	for _, function := range d.workload {
		ev, err = function(sources, period)
		if err != nil {
			return
		}
		events = append(events, ev...)
	}
	sort.Sort(events)
	return
}

func padding(s string, max int) string {
	count := max - len(s)
	for i := 0; i < count; i++ {
		s += " "
	}
	return s
}
