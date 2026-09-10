package rules

import (
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func TestCollectiveAblationsPreserveCostsAndChildProvisioning(t *testing.T) {
	for _, mode := range []string{"", "bonds", "sharing"} {
		t.Run(mode, func(t *testing.T) {
			w := fixture(t)
			w.Config.CollectiveAblation = mode
			p := w.Particles[1]
			act := func(op vm.Opcode, a int) {
				Resolve(w, Event{Actor: p.ID, Intent: vm.Intent{Instruction: vm.Instruction{Op: op, A: a}}})
			}
			act(vm.ALLOCATE, 1)
			q := w.Particles[p.Target]
			if q == nil || q.Energy == 0 {
				t.Fatal("child allocation/provisioning blocked")
			}
			before := q.Energy
			act(vm.TRANSFER, 5)
			if q.Energy != before+5 {
				t.Fatal("unbonded child transfer blocked")
			}
			energy := p.Energy
			act(vm.BIND, 0)
			if p.Energy != energy-1 || w.Linked(p.ID) != (mode != "bonds") {
				t.Fatal("bond ablation or instruction cost wrong")
			}
			before, energy = q.Energy, p.Energy
			act(vm.TRANSFER, 5)
			want := 5
			if mode == "sharing" {
				want = 0
			}
			if q.Energy != before+want || p.Energy != energy-1-want {
				t.Fatal("bonded transfer ablation or cost wrong")
			}
			check(t, w)
		})
	}
}

func TestSignalReadingAblationPreservesOtherChannelsAndEmission(t *testing.T) {
	w := fixture(t)
	w.Config.Environment = "coupled"
	w.Config.CollectiveAblation = "signal-reading"
	w.Environment = &world.EnvironmentState{}
	p := w.Particles[1]
	setProgram(w, p, []vm.Instruction{{Op: vm.EMIT, A: -1, B: 8}})
	before := p.Energy
	Resolve(w, Evaluate(w, p))
	if p.Energy != before-10 || w.Cells[p.Position].Signal != 8 || w.Environment.Emitted != 8 {
		t.Fatal("reading ablation altered emission")
	}
	for direction := 0; direction < 4; direction++ {
		setProgram(w, p, []vm.Instruction{{Op: vm.EMIT, A: direction, B: 8}})
		Resolve(w, Evaluate(w, p))
	}
	for _, channel := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15} {
		setProgram(w, p, []vm.Instruction{{Op: vm.SENSE, A: channel, B: 0}})
		w.Config.CollectiveAblation = ""
		expected := Evaluate(w, p).Sensed
		w.Config.CollectiveAblation = "signal-reading"
		if channel == 7 || channel >= 12 {
			expected = 0
		}
		Resolve(w, Evaluate(w, p))
		if p.Memory[0] != expected {
			t.Fatalf("channel %d changed incorrectly", channel)
		}
	}
	check(t, w)
}
