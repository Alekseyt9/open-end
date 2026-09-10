package rules

import (
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func symbolFixture(t *testing.T) (*world.World, *world.Particle) {
	w := fixture(t)
	w.Config.Environment = "coupled"
	w.Environment = &world.EnvironmentState{}
	w.Config.Symbols = "persistent"
	w.Symbols = &world.SymbolState{}
	p := w.Particles[1]
	setProgram(w, p, []vm.Instruction{{Op: vm.TOKEN, A: -1, B: 0}, {Op: vm.LISTEN, A: -1, B: 1}, {Op: vm.LOOKUP, A: 1, B: 2}})
	return w, p
}
func TestWordCompositionCostsExpiryAndTickStart(t *testing.T) {
	w, p := symbolFixture(t)
	before, _ := w.Totals()
	rng, transport := w.RNG, w.TransportRNG
	for _, token := range []int{0, 3, 1} {
		p.Memory[0] = token
		p.IP = 0
		Resolve(w, Evaluate(w, p))
	}
	after, _ := w.Totals()
	if before-after != 6 || w.Symbols.Writes != 3 || w.RNG != rng || w.TransportRNG != transport {
		t.Fatal("cost or RNG mismatch")
	}
	p.IP = 1
	e := Evaluate(w, p)
	if e.Sensed != 18 || e.WordLength != 2 || e.ForeignWord {
		t.Fatalf("ordered suffix [3,1] encoded incorrectly: %+v", e)
	}
	p.IP = 0
	p.Memory[0] = 2
	Resolve(w, Evaluate(w, p))
	Resolve(w, e)
	if p.Memory[1] != 18 || w.Symbols.PairReads != 1 {
		t.Fatal("read did not preserve tick-start word")
	}
	check(t, w)
	w.Tick = world.WordLifetime
	p.IP = 1
	if e := Evaluate(w, p); e.Sensed != 0 {
		t.Fatal("expired word readable")
	}
	Inflow(w)
	if w.Cells[p.Position].Word != nil {
		t.Fatal("expired word retained")
	}
	check(t, w)
}
func TestContextLookupHasNoBuiltInMeaningAndHandlesExtremeMemory(t *testing.T) {
	w, p := symbolFixture(t)
	for _, v := range []struct{ context, want int }{{0, 41}, {1, 99}} {
		p.Memory = [8]int{3, 0, v.context, 41, 99, 0, 0, 0}
		// Input register 0 contains the same received value in both contexts.
		e := Event{Actor: p.ID, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.LOOKUP, A: 0, B: 2}}}
		Resolve(w, e)
		if p.Memory[2] != v.want {
			t.Fatal("context did not select arbitrary memory entry")
		}
	}
	if w.Symbols.Lookups != 2 || w.Symbols.ContextLookups != 1 {
		t.Fatal("lookup counters")
	}
	p.Memory[0] = int(^uint(0) >> 1)
	p.Memory[2] = -p.Memory[0] - 1
	Resolve(w, Event{Actor: p.ID, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.LOOKUP, A: 0, B: 2}}})
	check(t, w)
}
func TestReceptionInterventionsPreserveCostsAndRNG(t *testing.T) {
	seen := map[int]bool{}
	for _, mode := range []string{"persistent", "scrambled", "unreadable"} {
		w, p := symbolFixture(t)
		w.Config.Symbols = mode
		p.Memory[0] = 2
		p.IP = 0
		Resolve(w, Evaluate(w, p))
		// A historical author may be dead; provenance must not depend on survival.
		w.NextID = 3
		w.Cells[p.Position].Word.Authors[0] = 2
		for tick := uint64(0); tick < 16; tick++ {
			w.Tick = tick
			p.IP = 1
			rng, energy := w.RNG, p.Energy
			e := Evaluate(w, p)
			Resolve(w, e)
			if w.RNG != rng || energy-p.Energy != 1 {
				t.Fatal("reception intervention changes cost/RNG")
			}
			if mode == "unreadable" && (e.Sensed != 0 || e.ForeignWord) {
				t.Fatal("unreadable word exposed")
			}
			if mode == "persistent" && (e.Sensed != 3 || !e.ForeignWord) {
				t.Fatal("persistent word changed")
			}
			if mode == "scrambled" {
				seen[e.Sensed] = true
				if e.WordLength != 1 || !e.ForeignWord {
					t.Fatal("scramble changed length/provenance")
				}
			}
		}
		check(t, w)
	}
	if len(seen) != 4 {
		t.Fatal("alphabet did not change across ticks")
	}
}
