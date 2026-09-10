package rules

import "open-end/internal/world"

// Diffusion is conservative neighbor mixing, on a separate RNG stream from
// mutation/execution. Interval 0 disables it for an explicit control run.
func transport(w *world.World) {
	matter := w.Config.MatterDiffusion > 0 && w.Tick%uint64(w.Config.MatterDiffusion) == 0
	chemical := w.Config.Ecology && w.Config.ChemicalDiffusion > 0 && w.Tick%uint64(w.Config.ChemicalDiffusion) == 0
	if !matter && !chemical {
		return
	}
	// Each field has its own round: unequal intervals must not lock one field
	// to a single axis. Interleave fields to preserve a stable resolution order.
	matterPhase, chemicalPhase := 0, 0
	if matter {
		matterPhase = int((w.Tick / uint64(w.Config.MatterDiffusion)) % 4)
	}
	if chemical {
		chemicalPhase = int((w.Tick / uint64(w.Config.ChemicalDiffusion)) % 4)
	}
	for pos := range w.Cells {
		if other, ok := mixNeighbor(w, pos, matterPhase); matter && ok {
			mix(w, &w.Cells[pos].Matter, &w.Cells[other].Matter)
		}
		if other, ok := mixNeighbor(w, pos, chemicalPhase); chemical && ok {
			for r := 0; r < 3; r++ {
				mix(w, &w.Cells[pos].Chemical[r], &w.Cells[other].Chemical[r])
			}
		}
	}
}

func mixNeighbor(w *world.World, pos, phase int) (int, bool) {
	coordinate, direction := pos%w.Config.Width, 1
	if phase >= 2 {
		coordinate, direction = pos/w.Config.Width, 2
	}
	return w.Neighbor(pos, direction), coordinate%2 == phase%2
}

func mix(w *world.World, a, b *int) {
	if *a == *b {
		return
	}
	if *a < *b {
		a, b = b, a
	}
	diff := *a - *b
	n := diff / 2
	if diff%2 != 0 {
		n += w.TransportRNG.Intn(2)
	}
	*a -= n
	*b += n
}

// X, Y, Z each represent one conserved unit of chemical matter; stored energy
// is 8, 4, 0 respectively. External field energy charges Z -> X, never creates
// chemical matter. There is no spontaneous Y source.
func charge(w *world.World) {
	for pos := range w.Cells {
		c := &w.Cells[pos]
		n := min(c.Chemical[2], c.Energy/8)
		c.Chemical[2] -= n
		c.Chemical[0] += n
		c.Energy -= n * 8
		w.Accounting.Charged += int64(n)
	}
}

// CONVERT selects a neutral local reaction, not an organism role.
// 0: X -> Y + 4 energy; 1: Y -> Z + 4 energy.
func convert(w *world.World, p *world.Particle, reaction, amount int) {
	if reaction < 0 || reaction > 1 || amount <= 0 {
		w.Accounting.FailedReaction++
		return
	}
	c := &w.Cells[p.Position]
	n := max(0, min(amount, c.Chemical[reaction], (w.Config.EnergyCapacity-p.Energy)/4))
	c.Chemical[reaction] -= n
	c.Chemical[reaction+1] += n
	p.Energy += 4 * n
	if n == 0 {
		w.Accounting.FailedReaction++
	}
	w.Accounting.Converted[reaction] += int64(n)
	if r := w.Genomes[p.Genome]; r != nil {
		r.Converted[reaction] += int64(n)
	}
}
