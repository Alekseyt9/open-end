package rules

import "open-end/internal/world"

// Observer receives value-only successful events; it is never serialized into
// the world or consulted by physics. Tick is the boundary after this action.
type Observer interface {
	Interaction(Interaction)
	Death(Death)
}
type Interaction struct {
	Tick           uint64
	Kind           string
	Source, Target string // genome hashes; empty means unprogrammed matter
	Energy         int64
	NewGenome      bool
}
type Death struct {
	Tick, Created uint64
	Genome        string
}

func emit(s Observer, w *world.World, kind string, p, q *world.Particle, energy int) {
	if s == nil {
		return
	}
	fresh := false
	if kind == "copy" {
		fresh = w.Genomes[q.Genome].Births == 1
	}
	s.Interaction(Interaction{Tick: w.Tick + 1, Kind: kind, Source: p.Genome, Target: q.Genome, Energy: int64(energy), NewGenome: fresh})
}
func unlinkObserved(s Observer, w *world.World, p *world.Particle) {
	if s != nil {
		seen := map[uint64]bool{}
		for d := 0; d < 4; d++ {
			id := w.Cells[w.Neighbor(p.Position, d)].Occupant
			if _, ok := w.Relations[world.RelationKey(p.ID, id)]; ok && !seen[id] {
				emit(s, w, "unbind", p, w.Particles[id], 0)
				seen[id] = true
			}
		}
	}
	w.Unlink(p.ID)
}
