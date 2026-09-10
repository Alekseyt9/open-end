package evolution

import (
	"open-end/internal/vm"
	"reflect"
	"testing"
)

func TestMutationBoundsAndInputOwnership(t *testing.T) {
	seed := vm.Seed()
	before := vm.Seed()
	r := RNG{State: 42}
	code := seed
	for i := 0; i < 10000; i++ {
		code = Mutate(code, &r, 1000000, 24)
		if len(code) < 1 || len(code) > 24 {
			t.Fatalf("length %d", len(code))
		}
		for _, ins := range code {
			if ins.Op >= vm.OpcodeCount {
				t.Fatal("invalid opcode")
			}
		}
	}
	if !reflect.DeepEqual(seed, before) {
		t.Fatal("mutated input")
	}
	clone := Mutate(seed, &r, 0, 24)
	clone[0].A++
	if !reflect.DeepEqual(seed, before) {
		t.Fatal("copy shares code storage")
	}
}

func TestRNGKnownSequence(t *testing.T) {
	r := RNG{}
	for _, want := range []uint64{0xe220a8397b1dcdaf, 0x6e789e6aa1b965f4, 0x06c45d188009454f} {
		if got := r.Uint64(); got != want {
			t.Fatalf("got %x want %x", got, want)
		}
	}
}
