package evolution

import (
	"open-end/internal/vm"
	"slices"
	"testing"
)

func TestPolicyEncoding(t *testing.T) {
	for a, want := range []int{10000, 0, 2500, 5000, 20000, 40000, 80000, 160000} {
		if got := PolicyRate(10000, a); got != want {
			t.Fatalf("rate %d: %d", a, got)
		}
		if PolicyRate(0, a) != 0 {
			t.Fatal("zero-mutation control changed")
		}
	}
	if PolicyRate(1000000, -1) != 1000000 {
		t.Fatal("clamp or negative indexing")
	}
	p := DecodePolicy(vm.Instruction{Op: vm.COPY, A: 4, B: 30}, 10000, true)
	if !p.Valid() || p.PPM != 5000 || p.Operator != 6 || !p.Recombine || !p.Proofread {
		t.Fatalf("%+v", p)
	}
	if fixed := DecodePolicy(vm.Instruction{Op: vm.COPY, A: 4, B: 30}, 10000, false); fixed != (CopyPolicy{Kind: "code", PPM: 10000}) {
		t.Fatalf("fixed policy %+v", fixed)
	}
	m := DecodePolicy(vm.Instruction{Op: vm.COPYMEM, A: 3, B: 6}, 10000, true)
	if !m.Valid() || m.PPM != 5000 || m.Layout != 2 || !m.Initial {
		t.Fatalf("%+v", m)
	}
}

func TestPolicyOperatorsAndRecombinationOwnership(t *testing.T) {
	seed := vm.Seed()
	before := slices.Clone(seed)
	for mode := 0; mode <= 6; mode++ {
		r := RNG{State: 42}
		code := slices.Clone(seed)
		for step := 0; step < 1000; step++ {
			previous := slices.Clone(code)
			code = MutatePolicy(code, &r, CopyPolicy{Kind: "code", PPM: 1000000, Operator: mode}, 24, int(vm.OpcodeCount))
			if len(code) < 1 || len(code) > 24 {
				t.Fatal("unbounded program")
			}
			for i, ins := range code {
				if ins.Op >= vm.OpcodeCount {
					t.Fatal("invalid opcode")
				}
				if mode == 6 && ins.Op != previous[i].Op {
					t.Fatal("operand edit changed opcode")
				}
			}
		}
	}
	r := RNG{State: 9}
	clone := MutatePolicy(seed, &r, CopyPolicy{PPM: 0}, 24, int(vm.OpcodeCount))
	if r.State != 9 || !slices.Equal(clone, seed) {
		t.Fatal("zero-rate consumed randomness")
	}
	clone[0].A++
	donor := []vm.Instruction{{Op: vm.COPY, A: 1, B: 6}}
	mixed := Recombine(seed, donor, &r)
	if len(mixed) != len(seed) || !slices.Contains(mixed, donor[0]) {
		t.Fatal("donor was not spliced")
	}
	for i := range mixed {
		mixed[i].B++
	}
	if !slices.Equal(seed, before) || donor[0].B != 6 {
		t.Fatal("aliased parent or donor")
	}
}

func TestInheritanceFormats(t *testing.T) {
	working := [8]int{1, 2, 3, 4, 5, 6, 7, 8}
	initial := [8]int{8, 7, 6, 5, 4, 3, 2, 1}
	for layout, want := range [][8]int{working, {1, 2, 3, 4}, {1, 0, 3, 0, 5, 0, 7}, {}} {
		r := RNG{State: 12}
		got := CopyMemory(working, initial, &r, CopyPolicy{Kind: "memory", Layout: layout})
		if got != want || r.State != 12 {
			t.Fatalf("layout %d: %v", layout, got)
		}
	}
	r := RNG{State: 12}
	if CopyMemory(working, initial, &r, CopyPolicy{Kind: "memory", Initial: true}) != initial {
		t.Fatal("initial source ignored")
	}
	for i := 0; i < 100; i++ {
		got := CopyMemory(working, initial, &r, CopyPolicy{Kind: "memory", PPM: 1000000, Layout: 2})
		for j := 1; j < 8; j += 2 {
			if got[j] != 0 {
				t.Fatal("mutation populated an untransmitted slot")
			}
		}
	}
}
