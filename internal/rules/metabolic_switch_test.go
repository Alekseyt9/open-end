package rules

import (
	"open-end/internal/vm"
	"testing"
)

func TestMetabolicSwitchCostsPreparationAndConservation(t *testing.T) {
	for _, cost := range []int{0, 1, 4} {
		w := fixture(t)
		w.Config.MetabolicSwitchCost = cost
		p := w.Particles[1]
		act := func(reaction, amount int) {
			Resolve(w, Event{Actor: 1, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.CONVERT, A: reaction, B: amount}}})
		}
		act(0, 4)
		act(0, 1)
		before := p.Energy
		act(1, 2)
		if p.Energy != before-1-cost+8 || w.Accounting.MetabolicSwitchEnergy != int64(cost) {
			t.Fatal("wrong switching charge")
		}
		want := 2
		if cost == 0 {
			want = 0
		}
		if p.MetabolicState != want {
			t.Fatal("wrong preparation state")
		}
		state, energy := p.MetabolicState, w.Accounting.MetabolicSwitchEnergy
		act(-1, 3)
		act(0, 0)
		if p.MetabolicState != state || w.Accounting.MetabolicSwitchEnergy != energy {
			t.Fatal("invalid reaction changed preparation")
		}
		check(t, w)
	}
}
func TestSwitchAttemptWithNoSubstrateAndNewbornReset(t *testing.T) {
	w := fixture(t)
	w.Config.MetabolicSwitchCost = 4
	p := w.Particles[1]
	// Reaction 1 prepares even with no Y. Repeating the preparation costs nothing.
	for range 2 {
		Resolve(w, Event{Actor: 1, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.CONVERT, A: 1, B: 4}}})
	}
	if p.MetabolicState != 2 || w.Accounting.MetabolicSwitches != 0 {
		t.Fatal("first preparation/retry wrong")
	}
	for _, i := range []vm.Instruction{{Op: vm.ALLOCATE, A: 1}, {Op: vm.COPYMEM}, {Op: vm.COPY}} {
		Resolve(w, Event{Actor: 1, Intent: vm.Intent{Instruction: i}})
	}
	if p.MetabolicState != 2 || w.Particles[p.Target].MetabolicState != 0 {
		t.Fatal("physiological state copied into newborn")
	}
	check(t, w)
}
func TestUnaffordableSwitchDissipatesReserveBeforeReaction(t *testing.T) {
	w := fixture(t)
	w.Config.MetabolicSwitchCost = 4
	p := w.Particles[1]
	p.MetabolicState = 1
	dissipate(w, p, p.Energy-4)
	chemical := w.Cells[p.Position].Chemical
	Resolve(w, Event{Actor: 1, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.CONVERT, A: 1, B: 4}}})
	if p.Energy != 0 || p.MetabolicState != 1 || w.Accounting.MetabolicSwitchEnergy != 3 || w.Accounting.MetabolicSwitchStarved != 1 || chemical != w.Cells[p.Position].Chemical {
		t.Fatal("switch starvation not conservative")
	}
	Decay(w, 1)
	check(t, w)
}
