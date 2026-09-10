package rules

import (
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

type motionSink struct{ moves, bonds, energy, unbinds int }

func (s *motionSink) Interaction(e Interaction) {
	if e.Kind == "unbind" {
		s.unbinds++
	}
}
func (s *motionSink) Death(Death) {}
func (s *motionSink) BondMoved(id uint64, bonds, energy int) {
	s.moves++
	s.bonds += bonds
	s.energy += energy
}
func motionFixture(t *testing.T) (*world.World, *world.Particle) {
	w := fixture(t)
	p := w.Particles[1]
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: 1}}})
	w.Relations[world.RelationKey(p.ID, p.Target)] = world.Relation{A: p.ID, B: p.Target}
	setProgram(w, p, []vm.Instruction{{Op: vm.MOVE, A: 0}})
	return w, p
}
func TestYieldingMotionConservesResourcesAndEmitsBreaks(t *testing.T) {
	for _, mode := range []string{"", "yielding"} {
		w, p := motionFixture(t)
		w.Config.BondMotion = mode
		pos, energy := p.Position, p.Energy
		before, matter := w.Totals()
		rng := w.RNG
		s := &motionSink{}
		ResolveObserved(w, Evaluate(w, p), s)
		want := 1
		if mode == "yielding" {
			want = 3
			if p.Position != w.Neighbor(pos, 0) || len(w.Relations) != 0 || s.moves != 1 || s.bonds != 1 || s.energy != 2 || s.unbinds != 1 {
				t.Fatal("rupture/movement not recorded")
			}
		} else if p.Position != pos || len(w.Relations) != 1 {
			t.Fatal("legacy bond movement changed")
		}
		after, m := w.Totals()
		if energy-p.Energy != want || before-after != int64(want) || matter != m || w.RNG != rng {
			t.Fatal("resource or RNG mismatch")
		}
		check(t, w)
	}
}
func TestFailedMovePreservesBondsAndOnlyPaysInstruction(t *testing.T) {
	for _, scenario := range []string{"occupied", "low-energy"} {
		w, p := motionFixture(t)
		w.Config.BondMotion = "yielding"
		if scenario == "occupied" {
			p.Code[0].A = 1
			setProgram(w, p, p.Code)
		} else {
			delta := p.Energy - 3
			p.Energy = 3
			w.Accounting.Dissipated += int64(delta)
		}
		pos, energy := p.Position, p.Energy
		s := &motionSink{}
		ResolveObserved(w, Evaluate(w, p), s)
		if p.Position != pos || len(w.Relations) != 1 || s.moves != 0 || s.unbinds != 0 || energy-p.Energy != 1 {
			t.Fatal("failed movement detached or charged rupture")
		}
		check(t, w)
	}
}
func TestDistinctBondCostOnTwoWideTorus(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 2
	c.Height = 3
	c.MaxEntities = 6
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	w.Config.BondMotion = "yielding"
	p := w.Particles[1]
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: 1}}})
	w.Relations[world.RelationKey(p.ID, p.Target)] = world.Relation{A: p.ID, B: p.Target}
	setProgram(w, p, []vm.Instruction{{Op: vm.MOVE, A: 0}})
	s := &motionSink{}
	ResolveObserved(w, Evaluate(w, p), s)
	if s.energy != 2 || s.bonds != 1 || s.unbinds != 1 {
		t.Fatal("same neighbor charged twice")
	}
	check(t, w)
}

func TestYieldPaysForEveryBrokenBond(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	w.Config.BondMotion = "yielding"
	for _, direction := range []int{1, 2, 3} {
		apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: direction}}})
		w.Relations[world.RelationKey(p.ID, p.Target)] = world.Relation{A: p.ID, B: p.Target}
	}
	setProgram(w, p, []vm.Instruction{{Op: vm.MOVE, A: 0}})
	energy := p.Energy
	s := &motionSink{}
	ResolveObserved(w, Evaluate(w, p), s)
	if s.bonds != 3 || s.energy != 6 || s.unbinds != 3 || energy-p.Energy != 7 || len(w.Relations) != 0 {
		t.Fatal("multiple rupture cost or events wrong")
	}
	check(t, w)
}

func TestBlockedRandomYieldDoesNotPerturbMutationRNG(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	w.Config.BondMotion = "yielding"
	for d := 0; d < 4; d++ {
		apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: d}}})
		w.Relations[world.RelationKey(p.ID, p.Target)] = world.Relation{A: p.ID, B: p.Target}
	}
	setProgram(w, p, []vm.Instruction{{Op: vm.MOVE, A: -1}})
	rng := w.RNG
	Resolve(w, Evaluate(w, p))
	if w.RNG != rng || len(w.Relations) != 4 {
		t.Fatal("blocked attempt changed RNG or bonds")
	}
	check(t, w)
}
