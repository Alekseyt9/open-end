package rules

import (
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func TestInterventionPreservesInstructionCostAndSkipsEffectRNG(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	setProgram(w, p, []vm.Instruction{{Op: vm.ALLOCATE, A: -1}, {Op: vm.NOP}})
	energy, rng, next := p.Energy, w.RNG, w.NextID
	calls := 0
	ResolveWithIntervention(w, Evaluate(w, p), nil, func(q *world.Particle, e *Event) bool { calls++; return true })
	if p.Energy != energy-4 || p.IP != 1 || w.RNG != rng || w.NextID != next || calls != 1 {
		t.Fatal("incorrect suppressed action")
	}
	check(t, w)
	// Instruction starvation occurs before the intervention callback.
	dissipate(w, p, p.Energy-1)
	ResolveWithIntervention(w, Evaluate(w, p), nil, func(q *world.Particle, e *Event) bool { calls++; return true })
	if calls != 1 || p.Energy != 0 {
		t.Fatal("starved action reached callback")
	}
}
