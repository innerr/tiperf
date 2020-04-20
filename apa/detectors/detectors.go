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

	// running status
	runnings map[string]bool
	found    FoundEvents
	result   FoundEvents
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
		make(map[string]bool),
		make(FoundEvents),
		make(FoundEvents),
	}
	d.RegisterAll()
	return d
}

func (d *Detectors) Register(name string, help string, function Detector, dependencies []string) {
	d.names = append(d.names, name)
	d.helps = append(d.helps, help)
	d.functions[name] = DetectorFunc{name, dependencies, function}
}

func (d *Detectors) RegisterCombined(name string, help string, names []string) {
	d.combinedNames = append(d.combinedNames, name)
	d.combinedHelps = append(d.combinedHelps, help)
	d.combineds[name] = names
}

// It's easier using z-index than using DAG to solve dependency,
//   since all detectors are statically configured
func (d *Detectors) RegisterAll() {
	d.Register("alive", "detect service up and down events", DetectAlive, []string{})

	d.Register("trend", "detect performance trend", DetectTrend, []string{"alive"})
	d.Register("balance", "detect anything imbalance", DetectBalance, []string{"alive"})

	d.Register("pikes", "detect performance pikes", DetectPikes, []string{"trend"})
	d.Register("jitter", "detect performance jitter", DetectJitter, []string{"trend"})

	d.RegisterCombined("all", "detect all", []string{
		"balance",
		"trend",
		"pikes",
		"alive",
		"jitter",
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

func (d *Detectors) ParseWorkloadFromArgs(args []string) (err error) {
	for _, arg := range args {
		err = d.ParseWorkloadFromArg(arg)
		if err != nil {
			return
		}
	}
	return
}

func (d *Detectors) ParseWorkloadFromArg(arg string) (err error) {
	if len(arg) == 0 {
		return
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
		return
	}

	name := arg
	function, ok := d.functions[name]
	if ok {
		d.workload[name] = function
		return
	}

	names, ok := d.combineds[name]
	if !ok {
		return fmt.Errorf("parsing workload, unknown name: " + name)
	}

	for _, it := range names {
		function, ok := d.functions[it]
		if !ok {
			return fmt.Errorf("unknown name: " + it + " in combination: " + name)
		}
		d.workload[it] = function
	}
	return
}

func (d *Detectors) GetWorkload() (workload []string) {
	for k, _ := range d.workload {
		workload = append(workload, k)
	}
	return
}

func (d *Detectors) RunWorkload(sources sources.Sources, period base.Period, con base.Console) (events Events, err error) {
	for name, _ := range d.workload {
		err = d.run(name, sources, period, con)
		if err != nil {
			return
		}
	}

	for _, it := range d.result {
		events = append(events, it...)
	}
	sort.Sort(events)

	d.found = FoundEvents{}
	d.result = FoundEvents{}
	if len(d.runnings) != 0 {
		panic("uncleaned running stack")
	}
	return
}

func (d *Detectors) run(name string, sources sources.Sources, period base.Period, con base.Console) (err error) {
	if _, ok := d.found[name]; ok {
		return
	}

	con.Debug("    ## detecting function ", name, " start\n")

	function, ok := d.functions[name]
	if !ok {
		return fmt.Errorf("function not found: " + name)
	}
	if _, ok := d.runnings[name]; ok {
		return fmt.Errorf("circled dependency: " + name)
	}
	d.runnings[name] = true

	for _, dependency := range function.Dependencies {
		err = d.run(dependency, sources, period, con)
		if err != nil {
			return
		}
	}

	events, err := function.Func(sources, period, d.found, con)
	if err != nil {
		return
	}
	sort.Sort(events)

	d.found[name] = events
	if _, ok := d.workload[name]; ok {
		d.result[name] = events
	} else {
		con.Debug("    ## detecting function ", name, " result is hidden\n")
	}
	delete(d.runnings, name)
	con.Debug("    ## detecting function ", name, " end\n")
	return
}

type DetectorFunc struct {
	Name         string
	Dependencies []string
	Func         Detector
}

type DetectorFuncs []DetectorFunc
