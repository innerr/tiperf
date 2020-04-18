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
	functions     map[string]DetectorFunc
	workload      map[string]DetectorFunc
}

func NewDetectors() Detectors {
	d := Detectors{
		make([]string, 0),
		make([]string, 0),
		make([]string, 0),
		make([]string, 0),
		make(map[string][]string),
		make(map[string]DetectorFunc),
		make(map[string]DetectorFunc),
	}
	d.RegisterAll()
	return d
}

func (d *Detectors) Register(name string, help string, function Detector, zIndex int) {
	d.names = append(d.names, name)
	d.helps = append(d.helps, help)
	d.functions[name] = DetectorFunc{name, zIndex, function}
}

func (d *Detectors) RegisterCombined(name string, help string, names []string) {
	d.combinedNames = append(d.combinedNames, name)
	d.combinedHelps = append(d.combinedHelps, help)
	d.combineds[name] = names
}

func (d *Detectors) RegisterAll() {
	d.Register("trend", "detect performance trend", DetectTrend, 0)
	d.Register("balance", "detect anything imbalance", DetectBalance, 1)
	d.Register("pikes", "detect performance pikes", DetectPikes, 10)
	d.Register("alive", "detect service up and down events", DetectAlive, 11)

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
			names, ok := d.combineds[name]
			if ok {
				for _, it := range names {
					delete(d.workload, it)
				}
			} else {
				delete(d.workload, name)
			}
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

func (d *Detectors) RunWorkload(sources sources.Sources, period base.Period) (events Events, err error) {
	funcs := make(DetectorFuncs, len(d.workload))
	i := 0
	for _, v := range d.workload {
		funcs[i] = v
		i++
	}
	sort.Sort(funcs)

	found := FoundEvents{}
	var es Events
	for _, function := range funcs {
		es, err = function.Func(sources, period, found)
		if err != nil {
			return
		}
		sort.Sort(es)
		found[function.Name] = es
		events = append(events, es...)
	}
	sort.Sort(events)
	return
}

// It's easier using z-index than using DAG to solve dependency,
//   since all detectors are statically configured
type DetectorFunc struct {
	Name   string
	ZIndex int
	Func   Detector
}

type DetectorFuncs []DetectorFunc

func (d DetectorFuncs) Len() int {
	return len(d)
}

func (d DetectorFuncs) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DetectorFuncs) Less(i, j int) bool {
	return d[i].ZIndex < d[j].ZIndex
}
