package rules

import (
	"open-end/internal/evolution"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func fixture(t *testing.T) *world.World {
	t.Helper()
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.Ecology = true
	c.MutationPPM = 0
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func setProgram(w *world.World, p *world.Particle, code []vm.Instruction) {
	p.Code = code
	p.IP = 0
	p.Origin = evolution.Hash(code, p.InitialMemory)
	w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
	w.RegisterGenome(p, "")
}

func check(t *testing.T, w *world.World) {
	t.Helper()
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestReactionBalancesAndNoFreeWaste(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	energy, matter := w.Totals()
	convert(w, p, 1, 8)
	if w.Accounting.Converted[1] != 0 {
		t.Fatal("Y was created without a source")
	}
	convert(w, p, 0, 4)
	convert(w, p, 1, 2)
	if w.Cells[p.Position].Chemical != [3]int{12, 2, 18} || p.Energy != 152 {
		t.Fatal("incorrect stoichiometry")
	}
	e, m := w.Totals()
	if e != energy || m != matter {
		t.Fatal("reaction creates resources")
	}
	charge(w)
	check(t, w)
}

func TestLimitedReactionAndInvalidAmounts(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	p.Energy = w.Config.EnergyCapacity - 2
	w.Accounting.InitialEnergy += int64(p.Energy - 128)
	convert(w, p, 0, 1000000)
	if w.Accounting.Converted[0] != 0 {
		t.Fatal("partial chemical unit was consumed at capacity")
	}
	for _, args := range [][2]int{{-1, 5}, {2, 5}, {0, -1}} {
		convert(w, p, args[0], args[1])
	}
	check(t, w)
}

func TestNeighborCanConsumeProducedResource(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: 1}}})
	q := w.Particles[p.Target]
	before := q.Energy
	convert(w, q, 1, 2)
	if q.Energy != before {
		t.Fatal("consumer used absent Y")
	}
	convert(w, p, 0, 8)
	// One conservative local exchange, the same primitive used by diffusion.
	mix(w, &w.Cells[p.Position].Chemical[1], &w.Cells[q.Position].Chemical[1])
	convert(w, q, 1, 2)
	if q.Energy != before+8 {
		t.Fatal("neighbor did not benefit from produced Y")
	}
	check(t, w)
}

func TestBondsBlockMovementAndReleaseOnDeath(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: 1}}})
	q := w.Particles[p.Target]
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.BIND}}})
	pos := p.Position
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.MOVE, A: 0}}})
	if p.Position != pos || len(w.Relations) != 1 {
		t.Fatal("bond did not retain adjacent matter")
	}
	// Repeated binding is idempotent.
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.BIND}}})
	if len(w.Relations) != 1 {
		t.Fatal("duplicate bond")
	}
	dissipate(w, q, q.Energy)
	Decay(w, q.ID)
	if len(w.Relations) != 0 {
		t.Fatal("dangling bond after death")
	}
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.MOVE, A: 0}}})
	if p.Position == pos {
		t.Fatal("particle stayed bound to dead neighbor")
	}
	check(t, w)
}

func TestTargetTakeAndUnbindAreLocal(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: 1}}})
	q := w.Particles[p.Target]
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.TARGET, A: 1}}})
	if p.Target != q.ID {
		t.Fatal("target selection")
	}
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.BIND}}})
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.UNBIND, A: -1}}})
	if len(w.Relations) != 0 {
		t.Fatal("unbind failed")
	}
	energy, _ := w.Totals()
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.TAKE, A: 100}}})
	if q.Energy != 0 || w.Accounting.Taken != 12 {
		t.Fatal("take exceeds victim resource")
	}
	e, _ := w.Totals()
	if e != energy {
		t.Fatal("theft creates energy")
	}
	Decay(w, q.ID)
	check(t, w)
	old := p.Energy
	apply(w, p, Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.TAKE, A: 100}}})
	if p.Energy != old {
		t.Fatal("stale target used")
	}
}

func TestDiffusionConservesAllFields(t *testing.T) {
	for _, dimensions := range [][2]int{{8, 8}, {7, 5}} {
		c := world.DefaultConfig()
		c.Width = dimensions[0]
		c.Height = dimensions[1]
		c.MaxEntities = c.Width * c.Height
		c.Ecology = true
		c.MatterDiffusion = 1
		c.ChemicalDiffusion = 4
		w, err := world.New(c)
		if err != nil {
			t.Fatal(err)
		}
		for i := range w.Cells {
			w.Cells[i].Chemical = [3]int{}
			w.Cells[i].Matter = 0
		}
		w.Cells[0].Chemical = [3]int{int(w.Accounting.InitialChemical) / 2, 0, int(w.Accounting.InitialChemical) / 2}
		w.Cells[0].Matter = len(w.Cells) - 1
		for i := 0; i < 200; i++ {
			transport(w)
			w.Tick++
			check(t, w)
		}
		rows, cols := map[int]bool{}, map[int]bool{}
		for pos, c := range w.Cells {
			if c.Chemical[0] > 0 {
				rows[pos/w.Config.Width] = true
				cols[pos%w.Config.Width] = true
			}
		}
		if len(rows) < 3 || len(cols) < 3 {
			t.Fatal("unequal intervals locked diffusion to an axis")
		}
	}
}

func TestEvaluateDoesNotMutateAndSensesTickStart(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	setProgram(w, p, []vm.Instruction{{Op: vm.SENSE, A: 3, B: 2}})
	e := Evaluate(w, p)
	if p.Memory[2] != 0 || e.Sensed != 16 {
		t.Fatal("evaluation changed state")
	}
	convert(w, p, 0, 4)
	Resolve(w, e)
	if p.Memory[2] != 16 {
		t.Fatal("sense reads effects of earlier resolutions")
	}
	check(t, w)
}
