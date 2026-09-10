package observer

import (
	"fmt"
	"open-end/internal/world"
	"sort"
)

type EnvironmentCell struct {
	Position int `json:"position"`
	Terrain  int `json:"terrain"`
	Signal   int `json:"signal"`
}
type EnvironmentMetrics struct {
	Model   string                 `json:"model"`
	Width   int                    `json:"width"`
	Height  int                    `json:"height"`
	Terrain int64                  `json:"terrain"`
	Signal  int64                  `json:"signal_energy"`
	State   world.EnvironmentState `json:"state"`
	Fields  []EnvironmentCell      `json:"fields"`
}
type EnvironmentActivity struct {
	Genome string `json:"genome"`
	world.EnvironmentActor
	Population int `json:"population_at_end"`
}
type EnvironmentWindow struct {
	Model              string                `json:"model"`
	Width              int                   `json:"width"`
	Height             int                   `json:"height"`
	FromTerrain        int64                 `json:"terrain_at_start"`
	Terrain            int64                 `json:"terrain_at_end"`
	Signal             int64                 `json:"signal_energy_at_end"`
	Built              int64                 `json:"built"`
	Reclaimed          int64                 `json:"reclaimed"`
	Eroded             int64                 `json:"eroded"`
	Emitted            int64                 `json:"emitted"`
	Decayed            int64                 `json:"decayed"`
	BlockedLight       int64                 `json:"blocked_light"`
	AttenuatedTransfer int64                 `json:"attenuated_transfer"`
	Sensed             uint64                `json:"sensed"`
	Fields             []EnvironmentCell     `json:"fields_at_end"`
	Activity           []EnvironmentActivity `json:"activity"`
}

func observeEnvironment(w *world.World) *EnvironmentMetrics {
	if w.Environment == nil {
		return nil
	}
	r := &EnvironmentMetrics{Model: w.Config.Environment, Width: w.Config.Width, Height: w.Config.Height, State: *w.Environment, Fields: []EnvironmentCell{}}
	if w.Environment.Actors != nil {
		r.State.Actors = map[string]*world.EnvironmentActor{}
		for hash, a := range w.Environment.Actors {
			copy := *a
			r.State.Actors[hash] = &copy
		}
	}
	for i, c := range w.Cells {
		r.Terrain += int64(c.Terrain)
		r.Signal += int64(c.Signal)
		if c.Terrain != 0 || c.Signal != 0 {
			r.Fields = append(r.Fields, EnvironmentCell{i, c.Terrain, c.Signal})
		}
	}
	return r
}
func validateEnvironmentMetrics(m *EnvironmentMetrics) error {
	if m.Model != "coupled" && m.Model != "inert" || m.Width < 2 || m.Height < 2 || m.Width > 1024 || m.Height > 1024 {
		return fmt.Errorf("invalid environment metadata")
	}
	// Reuse the physical ledger's conservation and actor checks on sparse fields.
	w := &world.World{Config: world.Config{Environment: m.Model}, Environment: &m.State, Cells: make([]world.Cell, m.Width*m.Height), Genomes: map[string]*world.GenomeRecord{}}
	for hash := range m.State.Actors {
		if len(hash) != 64 {
			return fmt.Errorf("invalid environment genome")
		}
		w.Genomes[hash] = &world.GenomeRecord{}
	}
	previous := -1
	var terrain, signal int64
	for _, c := range m.Fields {
		if c.Position <= previous || c.Position >= len(w.Cells) || c.Terrain == 0 && c.Signal == 0 {
			return fmt.Errorf("invalid environment field list")
		}
		w.Cells[c.Position].Terrain = c.Terrain
		w.Cells[c.Position].Signal = c.Signal
		terrain += int64(c.Terrain)
		signal += int64(c.Signal)
		previous = c.Position
	}
	if terrain != m.Terrain || signal != m.Signal {
		return fmt.Errorf("environment field totals disagree")
	}
	return w.ValidateEnvironment()
}
func summarizeEnvironment(frames []Metrics) (*EnvironmentWindow, error) {
	first, last := frames[0].Environment, frames[len(frames)-1].Environment
	if first == nil {
		for _, m := range frames {
			if m.Environment != nil {
				return nil, fmt.Errorf("environment appears within a session")
			}
		}
		return nil, nil
	}
	previous := first
	for _, m := range frames {
		e := m.Environment
		if e == nil || e.Model != first.Model || e.Width != first.Width || e.Height != first.Height {
			return nil, fmt.Errorf("inconsistent environment session")
		}
		if err := validateEnvironmentMetrics(e); err != nil {
			return nil, err
		}
		a, b := previous.State, e.State
		if b.Built < a.Built || b.Reclaimed < a.Reclaimed || b.Eroded < a.Eroded || b.Emitted < a.Emitted || b.Decayed < a.Decayed || b.BlockedLight < a.BlockedLight || b.AttenuatedTransfer < a.AttenuatedTransfer || b.Sensed < a.Sensed {
			return nil, fmt.Errorf("environment counters regressed")
		}
		for hash, old := range a.Actors {
			cur := b.Actors[hash]
			if cur == nil || cur.Built < old.Built || cur.Reclaimed < old.Reclaimed || cur.Emitted < old.Emitted || cur.Sensed < old.Sensed {
				return nil, fmt.Errorf("environment actor regressed")
			}
		}
		previous = e
	}
	a, b := first.State, last.State
	r := &EnvironmentWindow{Model: first.Model, Width: first.Width, Height: first.Height, FromTerrain: first.Terrain, Terrain: last.Terrain, Signal: last.Signal, Built: b.Built - a.Built, Reclaimed: b.Reclaimed - a.Reclaimed, Eroded: b.Eroded - a.Eroded, Emitted: b.Emitted - a.Emitted, Decayed: b.Decayed - a.Decayed, BlockedLight: b.BlockedLight - a.BlockedLight, AttenuatedTransfer: b.AttenuatedTransfer - a.AttenuatedTransfer, Sensed: b.Sensed - a.Sensed, Fields: append([]EnvironmentCell{}, last.Fields...), Activity: []EnvironmentActivity{}}
	pop := map[string]int{}
	for _, g := range frames[len(frames)-1].ActiveGenomes {
		pop[g.Hash] = g.Count
	}
	for hash, cur := range b.Actors {
		old := a.Actors[hash]
		if old == nil {
			old = &world.EnvironmentActor{}
		}
		d := world.EnvironmentActor{Built: cur.Built - old.Built, Reclaimed: cur.Reclaimed - old.Reclaimed, Emitted: cur.Emitted - old.Emitted, Sensed: cur.Sensed - old.Sensed}
		if d != (world.EnvironmentActor{}) {
			r.Activity = append(r.Activity, EnvironmentActivity{hash, d, pop[hash]})
		}
	}
	sort.Slice(r.Activity, func(i, j int) bool {
		a, b := r.Activity[i], r.Activity[j]
		if a.Built != b.Built {
			return a.Built > b.Built
		}
		return a.Genome < b.Genome
	})
	return r, nil
}
