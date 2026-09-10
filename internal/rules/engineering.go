package rules

import "open-end/internal/world"

// Negative direction selects the current cell; otherwise N/E/S/W modulo four.
func engineeringCell(w *world.World, p *world.Particle, direction int) int {
	if direction < 0 {
		return p.Position
	}
	return w.Neighbor(p.Position, direction)
}
func build(w *world.World, p *world.Particle, direction, amount int) {
	source := &w.Cells[p.Position]
	target := &w.Cells[engineeringCell(w, p, direction)]
	if amount > 0 {
		n := min(amount, source.Matter, 8-target.Terrain)
		if n == 0 {
			return
		}
		source.Matter -= n
		target.Terrain += n
		w.Environment.Built += int64(n)
		w.EnvironmentActor(p.Genome).Built += int64(n)
	} else if amount < 0 {
		// Clamp before negation, including the minimum signed integer operand.
		n := min(-max(amount, -8), target.Terrain)
		if n == 0 {
			return
		}
		target.Terrain -= n
		source.Matter += n
		w.Environment.Reclaimed += int64(n)
		w.EnvironmentActor(p.Genome).Reclaimed += int64(n)
	}
}
func emitSignal(w *world.World, p *world.Particle, direction, amount int) {
	c := &w.Cells[engineeringCell(w, p, direction)]
	n := max(0, min(amount, p.Energy-1, 64-c.Signal))
	if n == 0 {
		return
	}
	p.Energy -= n
	c.Signal += n
	w.Environment.Emitted += int64(n)
	w.EnvironmentActor(p.Genome).Emitted += int64(n)
}
func environmentStep(w *world.World) {
	// Stored construction matter returns locally. Signal energy becomes heat.
	for i := range w.Cells {
		c := &w.Cells[i]
		if w.Tick%64 == 0 && c.Terrain > 0 {
			c.Terrain--
			c.Matter++
			w.Environment.Eroded++
		}
		if w.Tick%8 == 0 && c.Signal > 0 {
			c.Signal--
			w.Environment.Decayed++
			w.Accounting.Dissipated++
		}
	}
	if w.Tick%4 != 0 {
		return
	}
	phase := int(w.Tick / 4 % 4)
	for i := range w.Cells {
		if j, ok := mixNeighbor(w, i, phase); ok {
			a, b := &w.Cells[i].Signal, &w.Cells[j].Signal
			n := (*a - *b) / 4
			*a -= n
			*b += n
		}
	}
}
func terrainMix(w *world.World, a, b *int, pos, other int) {
	if w.Config.Environment != "coupled" {
		mix(w, a, b)
		return
	}
	if *a < *b {
		a, b = b, a
	}
	diff := *a - *b
	if diff == 0 {
		return
	}
	n := diff / 2
	if diff%2 != 0 {
		n += w.TransportRNG.Intn(2)
	}
	reduced := n / (1 + max(w.Cells[pos].Terrain, w.Cells[other].Terrain))
	w.Environment.AttenuatedTransfer += int64(n - reduced)
	*a -= reduced
	*b += reduced
}
