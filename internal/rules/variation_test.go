package rules

import (
	"open-end/internal/evolution"
	"open-end/internal/vm"
	"testing"
)

func TestLocalRecombinationAndProofreadingCost(t *testing.T) {
	w := fixture(t)
	w.Config.CopyModel = "evolving"
	w.Config.MutationPPM = 1000000
	p := w.Particles[1]
	// Zero-rate donor creation; the actor is changed before recording any copy.
	setProgram(w, p, []vm.Instruction{{Op: vm.COPY, A: 1}})
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: 0}}})
	donor := w.Particles[p.Target]
	apply(w, p, Event{Intent: vm.Intent{Instruction: p.Code[0]}})
	// Keep the first actor's historical ledger valid; create another actor genome.
	setProgram(w, p, []vm.Instruction{{Op: vm.COPY, A: 0, B: 10}, {Op: vm.NOP}})
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: 1}}})
	child := w.Particles[p.Target]
	apply(w, p, Event{Intent: vm.Intent{Instruction: p.Code[0]}})
	record := w.Variation[evolution.PolicyKey(p.Genome, evolution.DecodePolicy(p.Code[0], 1000000, true))]
	if record == nil || record.Recombined != 1 || record.Donors[donor.Genome] != 1 || len(child.Code) == 0 {
		t.Fatal("local donor not recorded")
	}
	check(t, w)
	// An already coded target cannot count as another successful copy.
	apply(w, p, Event{Intent: vm.Intent{Instruction: p.Code[0]}})
	if record.Copies != 1 {
		t.Fatal("failed copy counted")
	}
	setProgram(w, p, []vm.Instruction{{Op: vm.COPY, A: 0, B: 16}})
	energy := p.Energy
	Resolve(w, Evaluate(w, p))
	if energy-p.Energy != 3 {
		t.Fatal("proofreading cost missing")
	}
	check(t, w)
}
