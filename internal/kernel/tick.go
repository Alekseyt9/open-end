// Package kernel controls deterministic scheduling and snapshot/replay.
package kernel

import (
	"open-end/internal/rules"
	"open-end/internal/world"
	"sort"
)

const Version = "0.1.0"
const RuleVersion = rules.Version

// Step executes at most one instruction per particle present at tick start.
// IDs are sorted then rotated by tick to avoid permanent ID priority. All
// intents are evaluated before effects; conflicts use this explicit order.
func Step(w *world.World) {
	rules.Inflow(w)
	ids := orderedIDs(w)
	events := make([]rules.Event, 0, len(ids))
	for j := range ids {
		id := ids[(j+int(w.Tick%uint64(len(ids))))%len(ids)]
		p := w.Particles[id]
		if len(p.Code) > 0 {
			events = append(events, rules.Evaluate(w, p))
		}
	}
	for _, e := range events {
		rules.Resolve(w, e)
	}
	// New allocations pay upkeep but first execute on the following tick.
	for _, id := range orderedIDs(w) {
		rules.Decay(w, id)
	}
	w.Tick++
}

func orderedIDs(w *world.World) []uint64 {
	ids := make([]uint64, 0, len(w.Particles))
	for id := range w.Particles {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
