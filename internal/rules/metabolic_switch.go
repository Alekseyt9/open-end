package rules

import "open-end/internal/world"

// Preparation is physiological state, not a cell type or copied genome trait.
// The first valid reaction is free. Switching attempts spend energy even when
// substrate is absent; an unaffordable switch spends the remaining reserve and
// leaves preparation unchanged. Ordinary instruction cost has already been paid.
func prepareReaction(w *world.World, p *world.Particle, reaction, amount int) bool {
	cost := w.Config.MetabolicSwitchCost
	if cost == 0 || reaction < 0 || reaction > 1 || amount <= 0 {
		return true
	}
	next := reaction + 1
	if p.MetabolicState != 0 && p.MetabolicState != next {
		spent := min(p.Energy, cost)
		w.Accounting.MetabolicSwitchEnergy += int64(spent)
		if p.Energy <= cost {
			dissipate(w, p, spent)
			w.Accounting.MetabolicSwitchStarved++
			return false
		}
		dissipate(w, p, spent)
		w.Accounting.MetabolicSwitches++
	}
	p.MetabolicState = next
	return true
}
