// Package adaptivity measures bounded causal proxies, never intrinsic intelligence.
package adaptivity

import (
	"bytes"
	"fmt"
	"math"
	"open-end/internal/kernel"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
)

const Version = 1

type Config struct {
	Ticks      int     `json:"ticks"`
	MinBenefit float64 `json:"minimum_mean_benefit"`
}

func DefaultConfig() Config { return Config{1000, .01} }
func (c Config) Validate() error {
	if c.Ticks < 30 || c.Ticks > 1000000 || math.IsNaN(c.MinBenefit) || c.MinBenefit < 0 || c.MinBenefit > 1 {
		return fmt.Errorf("invalid adaptivity config")
	}
	return nil
}

type Challenge struct {
	Name  string `json:"name"`
	Split string `json:"split"`
}

func Challenges() []Challenge {
	return []Challenge{{"dim", "selection"}, {"mix", "selection"}, {"pulse", "validation"}, {"mix-dim", "validation"}}
}

var Modes = []string{"intact", "frozen-perception", "memory-reset"}

type Probe struct {
	Challenge           string  `json:"challenge"`
	Split               string  `json:"split"`
	Mode                string  `json:"mode"`
	SourceHash          string  `json:"source_sha256"`
	InitialHash         string  `json:"initial_sha256"`
	FinalHash           string  `json:"final_sha256"`
	Ticks               int     `json:"ticks"`
	InitialExecutable   int     `json:"initial_executable"`
	FinalExecutable     int     `json:"final_executable"`
	PopulationRetention float64 `json:"mean_capped_population_retention"`
	ExecutableTrace     []int   `json:"executable_trace"`
	Copies              uint64  `json:"copies"`
	ChangedPerceptions  uint64  `json:"changed_perceptions"`
	ResetParticles      int     `json:"reset_particles"`
}

func Clone(source *world.World) (*world.World, error) {
	if source == nil {
		return nil, fmt.Errorf("nil source")
	}
	var b bytes.Buffer
	if err := kernel.Save(&b, source); err != nil {
		return nil, err
	}
	return kernel.Load(&b)
}
func Executable(w *world.World) int {
	n := 0
	for _, p := range w.Particles {
		if len(p.Code) > 0 {
			n++
		}
	}
	return n
}

// Counterfactual endpoints are reports only, not normal continuation snapshots.
func ProbeWorld(source *world.World, c Config, challenge Challenge, mode string) (Probe, error) {
	r := Probe{Challenge: challenge.Name, Split: challenge.Split, Mode: mode, Ticks: c.Ticks}
	if err := c.Validate(); err != nil {
		return r, err
	}
	valid := false
	for _, x := range Challenges() {
		if x == challenge {
			valid = true
		}
	}
	if !valid || mode != "intact" && mode != "frozen-perception" && mode != "memory-reset" {
		return r, fmt.Errorf("invalid probe")
	}
	w, err := Clone(source)
	if err != nil {
		return r, err
	}
	if source.RuleState != nil && len(source.RuleState.Pending) > 0 {
		return r, fmt.Errorf("probe requires no scheduled rule transitions")
	}
	r.SourceHash = kernel.Hash(source)
	r.InitialExecutable = Executable(source)
	if r.InitialExecutable == 0 {
		return r, fmt.Errorf("probe requires executable matter")
	}
	baseInflow := w.Config.Inflow
	switch challenge.Name {
	case "dim":
		w.Config.Inflow = baseInflow / 2
	case "pulse":
		w.Config.Inflow = 0
	case "mix", "mix-dim":
		// Permute free energy and chemistry without creating/removing resources.
		before := append([]world.Cell(nil), w.Cells...)
		dx, dy := w.Config.Width/2, 0
		if challenge.Name == "mix-dim" {
			dx = w.Config.Width / 3
			dy = w.Config.Height / 2
			w.Config.Inflow = baseInflow * 3 / 4
		}
		for i := range w.Cells {
			j := ((i/w.Config.Width+dy)%w.Config.Height)*w.Config.Width + (i%w.Config.Width+dx)%w.Config.Width
			w.Cells[i].Energy = before[j].Energy
			w.Cells[i].Chemical = before[j].Chemical
		}
	}
	if mode == "memory-reset" {
		for _, p := range w.Particles {
			if p.Memory != ([8]int{}) {
				r.ResetParticles++
				p.Memory = [8]int{}
			}
		}
	}
	if err := w.Validate(); err != nil {
		return r, err
	}
	r.InitialHash = kernel.Hash(w)
	type channel struct {
		ID uint64
		Op vm.Opcode
		A  int
	}
	type perceived struct {
		Value, Length int
		Foreign       bool
	}
	frozen := map[channel]perceived{}
	var filter func(*world.Particle, *rules.Event)
	if mode == "frozen-perception" {
		filter = func(p *world.Particle, e *rules.Event) {
			// Preserve self-energy sensing so disabling the seed's basic copying
			// threshold does not dominate this external-information assay.
			if e.Intent.Op == vm.SENSE && (e.Intent.A < 1 || e.Intent.A > 15) {
				return
			}
			if e.Intent.Op == vm.SENSE && e.Intent.A >= 6 && w.Config.Environment == "" {
				return
			}
			a := e.Intent.A
			if e.Intent.Op == vm.LISTEN {
				if a < 0 {
					a = -1
				} else {
					a = vm.Index(a, 4)
				}
			}
			key := channel{p.ID, e.Intent.Op, a}
			old, ok := frozen[key]
			if !ok {
				frozen[key] = perceived{e.Sensed, e.WordLength, e.ForeignWord}
				return
			}
			if e.Sensed != old.Value {
				r.ChangedPerceptions++
			}
			e.Sensed, e.WordLength, e.ForeignWord = old.Value, old.Length, old.Foreign
		}
	}
	for i := 0; i < c.Ticks; i++ {
		if challenge.Name == "pulse" && i == c.Ticks/3 {
			w.Config.Inflow = baseInflow
		}
		kernel.StepWithPerception(w, filter)
		n := Executable(w)
		r.ExecutableTrace = append(r.ExecutableTrace, n)
		r.PopulationRetention += math.Min(1, float64(n)/float64(r.InitialExecutable))
		if i%64 == 0 {
			for key := range frozen {
				if w.Particles[key.ID] == nil {
					delete(frozen, key)
				}
			}
		}
	}
	r.PopulationRetention /= float64(c.Ticks)
	r.FinalExecutable = Executable(w)
	r.Copies = w.Accounting.Copies - source.Accounting.Copies
	if err := w.Validate(); err != nil {
		return r, err
	}
	r.FinalHash = kernel.Hash(w)
	return r, nil
}

type Score struct {
	Robustness         float64 `json:"robustness"`
	PerceptionBenefit  float64 `json:"perception_benefit"`
	MemoryBenefit      float64 `json:"memory_benefit"`
	InformationBenefit float64 `json:"information_benefit"`
	AdaptiveProxy      float64 `json:"adaptive_proxy_0_100"`
	Productive         bool    `json:"productive"`
}

func ScoreProbes(probes []Probe, split string, c Config) (Score, error) {
	s := Score{}
	if err := c.Validate(); err != nil {
		return s, err
	}
	if split != "selection" && split != "validation" {
		return s, fmt.Errorf("invalid score split")
	}
	count := 0
	for _, ch := range Challenges() {
		if ch.Split != split {
			continue
		}
		arms := map[string]Probe{}
		for _, p := range probes {
			if p.Challenge == ch.Name {
				if p.Split != split || p.Ticks != c.Ticks || math.IsNaN(p.PopulationRetention) || p.PopulationRetention < 0 || p.PopulationRetention > 1 {
					return s, fmt.Errorf("invalid probe metrics")
				}
				if _, ok := arms[p.Mode]; ok {
					return s, fmt.Errorf("duplicate probe")
				}
				arms[p.Mode] = p
			}
		}
		if len(arms) != 3 {
			return s, fmt.Errorf("missing probe arms")
		}
		for _, mode := range Modes {
			if _, ok := arms[mode]; !ok {
				return s, fmt.Errorf("missing probe mode")
			}
		}
		a, b, d := arms[Modes[0]], arms[Modes[1]], arms[Modes[2]]
		if a.SourceHash != b.SourceHash || a.SourceHash != d.SourceHash || a.InitialExecutable != b.InitialExecutable || a.InitialExecutable != d.InitialExecutable {
			return s, fmt.Errorf("unmatched probes")
		}
		s.Robustness += a.PopulationRetention
		s.PerceptionBenefit += a.PopulationRetention - b.PopulationRetention
		s.MemoryBenefit += a.PopulationRetention - d.PopulationRetention
		s.Productive = s.Productive || a.Copies > 0
		count++
	}
	if count == 0 {
		return s, fmt.Errorf("empty split")
	}
	s.Robustness /= float64(count)
	s.PerceptionBenefit /= float64(count)
	s.MemoryBenefit /= float64(count)
	s.InformationBenefit = (math.Max(0, s.PerceptionBenefit) + math.Max(0, s.MemoryBenefit)) / 2
	if s.Productive && s.InformationBenefit >= c.MinBenefit {
		s.AdaptiveProxy = 100 * s.Robustness * s.InformationBenefit
	}
	return s, nil
}
