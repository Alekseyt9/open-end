package experiment

import (
	"bytes"
	"fmt"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
)

type ClonalCandidate struct {
	Cohort    int                `json:"cohort"`
	Members   []uint64           `json:"members"`
	Since     uint64             `json:"since_tick"`
	Qualified uint64             `json:"qualified_tick"`
	Last      uint64             `json:"last_observed_tick"`
	Founder   uint64             `json:"founder"`
	Genomes   []observer.Lineage `json:"genomes_at_qualification"`
}
type MulticellDiagnostics struct {
	InitialLinked         int               `json:"initial_linked_particles"`
	InitialMoveCode       int               `json:"initial_linked_with_move_code"`
	InitialUnbindCode     int               `json:"initial_linked_with_unbind_code"`
	InitialGrowthFrontier int               `json:"initial_linked_with_free_matter_neighbor"`
	LinkedMoveIntents     uint64            `json:"tick_start_linked_move_intents"`
	LinkedAllocateIntents uint64            `json:"tick_start_linked_allocate_intents"`
	NoSpace               uint64            `json:"tick_start_allocation_no_space"`
	NoMatter              uint64            `json:"tick_start_allocation_no_matter"`
	LowReserve            uint64            `json:"tick_start_allocation_low_reserve"`
	LinkedAllocations     uint64            `json:"successful_linked_allocations"`
	LinkedCopies          uint64            `json:"successful_linked_copies"`
	YieldMoves            uint64            `json:"successful_yield_moves"`
	BrokenBonds           uint64            `json:"motion_broken_bonds"`
	BreakEnergy           uint64            `json:"bond_break_energy"`
	DescendantGroupTicks  uint64            `json:"founder_free_descendant_group_ticks"`
	SingleFounderTicks    uint64            `json:"single_founder_group_ticks"`
	MultiFounderTicks     uint64            `json:"multiple_founder_group_ticks"`
	ParentAbsentTicks     uint64            `json:"parent_absent_group_ticks"`
	Skipped               uint64            `json:"skipped_clonal_group_ticks"`
	MaxConcurrentClonal   int               `json:"max_concurrent_qualified_clonal_groups"`
	QualifiedClonalTicks  uint64            `json:"qualified_clonal_group_ticks"`
	Clonal                []ClonalCandidate `json:"clonal_daughter_candidates"`
}
type MulticellTrial struct {
	EnvironmentTrial
	Diagnostics MulticellDiagnostics `json:"diagnostics"`
}
type multicellSink struct {
	*observer.Tracker
	w        *world.World
	d        MulticellDiagnostics
	minAge   uint64
	pending  map[string]uint64
	credited map[string]int
}

func freeSites(w *world.World, p *world.Particle, direction int) (space, matter bool) {
	for d := 0; d < 4; d++ {
		if direction >= 0 && vm.Index(direction, 4) != d {
			continue
		}
		c := w.Cells[w.Neighbor(p.Position, d)]
		space = space || c.Occupant == 0
		matter = matter || c.Occupant == 0 && c.Matter > 0
	}
	return
}
func (s *multicellSink) initial() {
	for _, p := range s.w.Particles {
		if !s.w.Linked(p.ID) {
			continue
		}
		s.d.InitialLinked++
		move, unbind := false, false
		for _, i := range p.Code {
			move = move || i.Op == vm.MOVE
			unbind = unbind || i.Op == vm.UNBIND
		}
		if move {
			s.d.InitialMoveCode++
		}
		if unbind {
			s.d.InitialUnbindCode++
		}
		if _, matter := freeSites(s.w, p, -1); matter {
			s.d.InitialGrowthFrontier++
		}
	}
}
func (s *multicellSink) inspect() {
	for _, p := range s.w.Particles {
		if len(p.Code) == 0 || !s.w.Linked(p.ID) {
			continue
		}
		i := vm.Decode(p.Code, p.IP, p.Flag)
		if i.Op == vm.MOVE {
			s.d.LinkedMoveIntents++
		}
		if i.Op == vm.ALLOCATE {
			s.d.LinkedAllocateIntents++
			space, matter := freeSites(s.w, p, i.A)
			if !space {
				s.d.NoSpace++
			} else if !matter {
				s.d.NoMatter++
			}
			if p.Energy <= 16 {
				s.d.LowReserve++
			}
		}
	}
}
func (s *multicellSink) Interaction(e rules.Interaction) {
	if s.w.Linked(e.SourceID) {
		if e.Kind == "allocate" {
			s.d.LinkedAllocations++
		}
		if e.Kind == "copy" {
			s.d.LinkedCopies++
		}
	}
	s.Tracker.Interaction(e)
}
func (s *multicellSink) BondMoved(id uint64, bonds, energy int) {
	s.d.YieldMoves++
	s.d.BrokenBonds += uint64(bonds)
	s.d.BreakEnergy += uint64(energy)
}
func (s *multicellSink) TickCompleted(tick uint64) {
	s.Tracker.TickCompleted(tick)
	s.recordBranches(tick, s.DescendantBranches())
}
func (s *multicellSink) recordBranches(tick uint64, branches []observer.DescendantBranch) {
	eligible := map[string]bool{}
	concurrent := 0
	for _, b := range branches {
		s.d.DescendantGroupTicks++
		if b.Contributors == 1 {
			s.d.SingleFounderTicks++
		} else {
			s.d.MultiFounderTicks++
		}
		if !b.ParentPresent {
			s.d.ParentAbsentTicks++
		}
		if b.Contributors != 1 || !b.ParentPresent {
			continue
		}
		key := fmt.Sprintf("%d:%v", b.Cohort, b.Members)
		eligible[key] = true
		since, ok := s.pending[key]
		if !ok {
			since = tick
			s.pending[key] = tick
		}
		if tick-since < s.minAge {
			continue
		}
		concurrent++
		s.d.QualifiedClonalTicks++
		if index, ok := s.credited[key]; ok {
			s.d.Clonal[index].Last = tick
			continue
		}
		if len(s.d.Clonal) >= 1024 {
			s.d.Skipped++
			continue
		}
		s.credited[key] = len(s.d.Clonal)
		s.d.Clonal = append(s.d.Clonal, ClonalCandidate{Cohort: b.Cohort, Members: b.Members, Since: since, Qualified: tick, Last: tick, Founder: b.FounderIDs[0], Genomes: b.Genomes})
	}
	s.d.MaxConcurrentClonal = max(s.d.MaxConcurrentClonal, concurrent)
	for key := range s.pending {
		if !eligible[key] {
			delete(s.pending, key)
		}
	}
}

// ContinueMulticell keeps the original daughter criterion and separately tracks
// stable one-founder clonal branches. Motion is the only physical intervention.
func ContinueMulticell(source *world.World, mode string, ticks, every int, minAge uint64) (*world.World, []observer.Metrics, MulticellTrial, error) {
	r := MulticellTrial{EnvironmentTrial: EnvironmentTrial{Variant: mode}}
	if source == nil || source.Config.BondMotion != "" || source.Config.CollectiveAblation != "" || (mode != "intact" && mode != "yielding") || ticks < 1 || every < 1 || minAge == 0 {
		return nil, nil, r, fmt.Errorf("invalid multicell continuation")
	}
	var buf bytes.Buffer
	if err := kernel.Save(&buf, source); err != nil {
		return nil, nil, r, err
	}
	w, err := kernel.Load(&buf)
	if err != nil {
		return nil, nil, r, err
	}
	if mode == "yielding" {
		w.Config.BondMotion = mode
	}
	r.SourceHash = kernel.Hash(source)
	r.InitialHash = kernel.Hash(w)
	r.Seed = w.Config.Seed
	s := &multicellSink{Tracker: observer.NewTracker(w), w: w, minAge: minAge, pending: map[string]uint64{}, credited: map[string]int{}, d: MulticellDiagnostics{Clonal: []ClonalCandidate{}}}
	if err := s.EnableCollectives(w, source, minAge); err != nil {
		return nil, nil, r, err
	}
	s.initial()
	frames := []observer.Metrics{s.Frame(w)}
	for i := 1; i <= ticks; i++ {
		s.inspect()
		kernel.StepObserved(w, s)
		if i%every == 0 || i == ticks {
			frames = append(frames, s.Frame(w))
		}
	}
	if err := w.Validate(); err != nil {
		return nil, nil, r, err
	}
	r.Summary, err = observer.Summarize(frames, uint64(ticks))
	r.FinalHash = kernel.Hash(w)
	r.Diagnostics = s.d
	return w, frames, r, err
}
