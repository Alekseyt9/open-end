package experiment

import (
	"bytes"
	"fmt"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/rules"
	"open-end/internal/world"
	"slices"
	"strconv"
)

type MetabolicProfile struct {
	ID     uint64   `json:"id"`
	Genome string   `json:"genome"`
	Units  [2]int64 `json:"reaction_units"`
	Copies uint64   `json:"copies"`
	Alive  bool     `json:"alive_at_window_end"`
	Class  string   `json:"class"`
}
type MetabolicWindow struct {
	From, To            uint64
	Population          int                `json:"population"`
	Units               [2]int64           `json:"reaction_units"`
	Switches            uint64             `json:"switches"`
	SwitchEnergy        int64              `json:"switch_energy"`
	SwitchStarved       uint64             `json:"switch_starved"`
	Specialists         [2]int             `json:"living_specialists"`
	Persistent          [2]int             `json:"same_cell_specialists_in_consecutive_windows"`
	Generalists         int                `json:"living_generalists"`
	Profiles            []MetabolicProfile `json:"active_profiles"`
	ComplementaryGroups [][]uint64         `json:"complementary_bonded_groups"`
}
type SwitchTrial struct {
	EnvironmentTrial
	WindowTicks   int               `json:"window_ticks"`
	MinimumUnits  int               `json:"minimum_profile_units"`
	PurityPercent int               `json:"specialist_purity_percent"`
	Windows       []MetabolicWindow `json:"metabolic_windows"`
}
type switchSink struct {
	*observer.Tracker
	w          *world.World
	from       uint64
	accounting world.Accounting
	cells      map[uint64]*MetabolicProfile
	previous   map[uint64]string
}

func (s *switchSink) entry(id uint64) *MetabolicProfile {
	p := s.cells[id]
	if p == nil {
		p = &MetabolicProfile{ID: id, Genome: s.w.Particles[id].Genome}
		s.cells[id] = p
	}
	return p
}
func (s *switchSink) ChemicalMoved(int, int, int, int, int) {}
func (s *switchSink) Reacted(e rules.ReactionEvent) {
	if e.Reaction < 0 || e.Reaction > 1 {
		return
	}
	n := e.Before[e.Reaction] - e.After[e.Reaction]
	if n > 0 {
		s.entry(e.Actor).Units[e.Reaction] += int64(n)
	}
}
func (s *switchSink) Interaction(e rules.Interaction) {
	if e.Kind == "copy" {
		s.entry(e.SourceID).Copies++
	}
	s.Tracker.Interaction(e)
}
func profileClass(u [2]int64) string {
	total := u[0] + u[1]
	if total < 32 {
		return "low-activity"
	}
	if u[0]*10 >= total*9 {
		return "reaction0"
	}
	if u[1]*10 >= total*9 {
		return "reaction1"
	}
	return "generalist"
}
func (s *switchSink) flush() MetabolicWindow {
	a := s.w.Accounting
	r := MetabolicWindow{From: s.from, To: s.w.Tick, Population: len(s.w.Particles), Switches: a.MetabolicSwitches - s.accounting.MetabolicSwitches, SwitchEnergy: a.MetabolicSwitchEnergy - s.accounting.MetabolicSwitchEnergy, SwitchStarved: a.MetabolicSwitchStarved - s.accounting.MetabolicSwitchStarved, Profiles: []MetabolicProfile{}, ComplementaryGroups: [][]uint64{}}
	ids := make([]uint64, 0, len(s.cells))
	for id := range s.cells {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	living := map[uint64]string{}
	for _, id := range ids {
		p := s.cells[id]
		for j := range 2 {
			r.Units[j] += p.Units[j]
		}
		p.Class = profileClass(p.Units)
		p.Alive = s.w.Particles[id] != nil
		if p.Class == "low-activity" {
			continue
		}
		r.Profiles = append(r.Profiles, *p)
		if !p.Alive {
			continue
		}
		living[id] = p.Class
		if p.Class == "generalist" {
			r.Generalists++
			continue
		}
		j := 0
		if p.Class == "reaction1" {
			j = 1
		}
		r.Specialists[j]++
		if s.previous[id] == p.Class {
			r.Persistent[j]++
		}
	}
	groups, _ := BondComponents(s.w)
	for _, group := range groups {
		has0, has1 := false, false
		for _, id := range group {
			has0 = has0 || living[id] == "reaction0"
			has1 = has1 || living[id] == "reaction1"
		}
		if has0 && has1 {
			r.ComplementaryGroups = append(r.ComplementaryGroups, group)
		}
	}
	s.previous = living
	s.cells = map[uint64]*MetabolicProfile{}
	s.accounting = a
	s.from = s.w.Tick
	return r
}

// ContinueSwitch introduces a recorded physical parameter without rewriting any
// genome. Zero cost retains legacy state, physics and serialization exactly.
func ContinueSwitch(source *world.World, mode string, ticks, every int, minAge uint64) (*world.World, []observer.Metrics, SwitchTrial, error) {
	cost, err := strconv.Atoi(mode)
	r := SwitchTrial{EnvironmentTrial: EnvironmentTrial{Variant: mode}, WindowTicks: min(5000, ticks), MinimumUnits: 32, PurityPercent: 90, Windows: []MetabolicWindow{}}
	if err != nil || cost < 0 || cost > 64 || source == nil || !source.Config.Ecology || source.RuleState != nil || source.Config.MetabolicSwitchCost != 0 || source.Config.BondMotion != "" || source.Config.CollectiveAblation != "" || ticks < 1 || every < 1 || minAge == 0 {
		return nil, nil, r, fmt.Errorf("invalid switching continuation")
	}
	var buf bytes.Buffer
	if err := kernel.Save(&buf, source); err != nil {
		return nil, nil, r, err
	}
	w, err := kernel.Load(&buf)
	if err != nil {
		return nil, nil, r, err
	}
	w.Config.MetabolicSwitchCost = cost
	r.SourceHash = kernel.Hash(source)
	r.InitialHash = kernel.Hash(w)
	r.Seed = w.Config.Seed
	s := &switchSink{Tracker: observer.NewTracker(w), w: w, from: w.Tick, accounting: w.Accounting, cells: map[uint64]*MetabolicProfile{}, previous: map[uint64]string{}}
	if err := s.EnableCollectives(w, source, minAge); err != nil {
		return nil, nil, r, err
	}
	frames := []observer.Metrics{s.Frame(w)}
	for i := 1; i <= ticks; i++ {
		kernel.StepObserved(w, s)
		if i%r.WindowTicks == 0 || i == ticks {
			r.Windows = append(r.Windows, s.flush())
		}
		if i%every == 0 || i == ticks {
			frames = append(frames, s.Frame(w))
		}
	}
	if err := w.Validate(); err != nil {
		return nil, nil, r, err
	}
	r.Summary, err = observer.Summarize(frames, uint64(ticks))
	r.FinalHash = kernel.Hash(w)
	return w, frames, r, err
}
