// Package kernel controls deterministic scheduling and snapshot/replay.
package kernel

import (
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"sort"
)

const Version = "0.1.0"
const RuleVersion = rules.Version

type Observer interface {
	rules.Observer
	TickCompleted(uint64)
}

// Step executes at most one instruction per particle present at tick start.
// IDs are sorted then rotated by tick to avoid permanent ID priority. All
// intents are evaluated before effects; conflicts use this explicit order.
func Step(w *world.World) {
	StepObserved(w, nil)
}

func StepObserved(w *world.World, sink Observer) {
	step(w, sink, nil, nil)
}

// StepWithPerception is an explicit experimental intervention, not observation.
// The caller must record/replay the filter protocol; snapshots do not persist it.
func StepWithPerception(w *world.World, filter func(*world.Particle, *rules.Event)) {
	step(w, nil, filter, nil)
}

// StepWithIntervention applies an explicitly recorded assay protocol. Snapshot
// state alone does not preserve the callback or its lineage tags.
func StepWithIntervention(w *world.World, sink Observer, beforeApply func(*world.Particle, *rules.Event) bool) {
	step(w, sink, nil, beforeApply)
}

func step(w *world.World, sink Observer, filter func(*world.Particle, *rules.Event), beforeApply func(*world.Particle, *rules.Event) bool) {
	if s := w.RuleState; s != nil && len(s.Pending) > 0 && s.Pending[0].Tick == w.Tick {
		change := s.Pending[0]
		s.Pending = s.Pending[1:]
		s.Apply(change)
	}
	rules.Inflow(w)
	ids := orderedIDs(w)
	events := make([]rules.Event, 0, len(ids))
	for j := range ids {
		id := ids[(j+int(w.Tick%uint64(len(ids))))%len(ids)]
		p := w.Particles[id]
		if len(p.Code) > 0 {
			e := rules.Evaluate(w, p)
			if filter != nil && (e.Intent.Op == vm.SENSE || e.Intent.Op == vm.LISTEN) {
				filter(p, &e)
			}
			events = append(events, e)
		}
	}
	for _, e := range events {
		rules.ResolveWithIntervention(w, e, sink, beforeApply)
	}
	// New allocations pay upkeep but first execute on the following tick.
	for _, id := range orderedIDs(w) {
		rules.DecayObserved(w, id, sink)
	}
	w.Tick++
	if sink != nil {
		sink.TickCompleted(w.Tick)
	}
}

func orderedIDs(w *world.World) []uint64 {
	ids := make([]uint64, 0, len(w.Particles))
	for id := range w.Particles {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
