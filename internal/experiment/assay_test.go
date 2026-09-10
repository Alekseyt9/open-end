package experiment

import (
	"open-end/internal/evolution"
	"open-end/internal/vm"
	"testing"
)

func testPair() Pair {
	a, b := vm.EcologySeed(), vm.EcologySeed()
	b[8] = vm.Instruction{Op: vm.BIND}
	return Pair{A: Template{evolution.Hash(a, [8]int{}), a}, B: Template{evolution.Hash(b, [8]int{}), b}}
}

func TestMatchedInoculation(t *testing.T) {
	pair := testPair()
	mixed, err := Inoculate(pair, 7, "mixed", 32)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{"a-half", "b-half", "a-only", "b-only"} {
		w, err := Inoculate(pair, 7, c, 32)
		if err != nil {
			t.Fatal(err)
		}
		want := 32
		if c == "a-half" || c == "b-half" {
			want = 16
		}
		if len(w.Particles) != want || w.Accounting.Copies != 0 {
			t.Fatal("incorrect founder counts")
		}
		for _, p := range w.Particles {
			if mixed.Cells[p.Position].Occupant == 0 {
				t.Fatal("unmatched initial position")
			}
			if p.Memory != [8]int{} || p.Energy != 128 {
				t.Fatal("unstandardized founder")
			}
		}
		if w.RNG != mixed.RNG || w.TransportRNG != mixed.TransportRNG {
			t.Fatal("placement consumed physics RNG")
		}
	}
}
