package rules

import "open-end/internal/world"

// BondMotionObserver records mechanical work without entering physical state.
type BondMotionObserver interface {
	BondMoved(id uint64, bonds, energy int)
}

// A particle pays two additional energy units per distinct bond only when it
// can enter a free cell and retain positive energy after breaking those bonds.
// The ordinary MOVE instruction cost has already been paid by Resolve.
func moveYielding(w *world.World, p *world.Particle, direction int, sink Observer) {
	bonds := map[uint64]bool{}
	for d := 0; d < 4; d++ {
		id := w.Cells[w.Neighbor(p.Position, d)].Occupant
		if _, ok := w.Relations[world.RelationKey(p.ID, id)]; ok {
			bonds[id] = true
		}
	}
	cost := 2 * len(bonds)
	if p.Energy <= cost {
		return
	}
	// A completely blocked attempt must not consume a random direction: doing
	// so would perturb later mutation draws despite no mechanical effect.
	space := false
	for d := 0; d < 4; d++ {
		if direction >= 0 && (direction%4+4)%4 != d {
			continue
		}
		space = space || w.Cells[w.Neighbor(p.Position, d)].Occupant == 0
	}
	if !space {
		return
	}
	pos := destination(w, p, direction, false)
	if pos < 0 {
		return
	}
	dissipate(w, p, cost)
	unlinkObserved(sink, w, p)
	w.Cells[p.Position].Occupant = 0
	w.Cells[pos].Occupant = p.ID
	p.Position = pos
	if g := w.Genomes[p.Genome]; g != nil {
		g.Moved++
	}
	if s, ok := sink.(BondMotionObserver); ok {
		s.BondMoved(p.ID, len(bonds), cost)
	}
}
