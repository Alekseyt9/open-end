package evolution

import (
	"fmt"
	"slices"

	"open-end/internal/vm"
)

// CopyPolicy is encoded in ordinary inherited instruction operands.
type CopyPolicy struct {
	Kind      string `json:"kind"`
	PPM       int    `json:"ppm"`
	Operator  int    `json:"operator"`
	Recombine bool   `json:"recombine"`
	Proofread bool   `json:"proofread"`
	Layout    int    `json:"layout"`
	Initial   bool   `json:"initial_memory"`
}
type CopyRecord struct {
	Genome     string            `json:"genome"`
	Policy     CopyPolicy        `json:"policy"`
	Copies     uint64            `json:"copies"`
	Changed    uint64            `json:"changed"`
	Recombined uint64            `json:"recombined"`
	Donors     map[string]uint64 `json:"donors,omitempty"`
	FirstTick  uint64            `json:"first_tick"`
	LastTick   uint64            `json:"last_tick"`
}

func (p CopyPolicy) Valid() bool {
	if p.PPM < 0 || p.PPM > 1000000 {
		return false
	}
	switch p.Kind {
	case "code":
		return p.Operator >= 0 && p.Operator <= 6 && p.Layout == 0 && !p.Initial
	case "memory":
		return p.Operator == 0 && !p.Recombine && !p.Proofread && p.Layout >= 0 && p.Layout < 4
	}
	return false
}
func PolicyKey(genome string, p CopyPolicy) string {
	return fmt.Sprintf("%s/%s/%d/%d/%t/%t/%d/%t", genome, p.Kind, p.PPM, p.Operator, p.Recombine, p.Proofread, p.Layout, p.Initial)
}
func PolicyRate(base, operand int) int {
	switch vm.Index(operand, 8) {
	case 1:
		return 0
	case 2:
		return base / 4
	case 3:
		return base / 2
	case 4:
		return min(1000000, base*2)
	case 5:
		return min(1000000, base*4)
	case 6:
		return min(1000000, base*8)
	case 7:
		return min(1000000, base*16)
	}
	return base
}
func DecodePolicy(i vm.Instruction, base int, evolving bool) CopyPolicy {
	a, b := i.A, i.B
	if !evolving {
		a, b = 0, 0
	}
	p := CopyPolicy{Kind: "code", PPM: PolicyRate(base, a)}
	if i.Op == vm.COPYMEM {
		p.Kind = "memory"
		p.Layout = vm.Index(b, 4)
		p.Initial = vm.Index(b, 8)&4 != 0
	} else {
		mode := vm.Index(b, 8)
		if mode == 7 {
			mode = 0
		}
		p.Operator = mode
		p.Recombine = vm.Index(b, 16)&8 != 0
		p.Proofread = vm.Index(b, 32)&16 != 0
		if p.Proofread {
			p.PPM /= 4
		}
	}
	return p
}

// MutatePolicy supports six bounded edits. Mode 0 samples all six; modes 1..6
// select replacement, insertion, deletion, duplication, segment deletion,
// and an operand edit respectively. It never aliases either input slice.
func MutatePolicy(code []vm.Instruction, r *RNG, p CopyPolicy, limit, opcodes int) []vm.Instruction {
	out := slices.Clone(code)
	if len(out) == 0 || !r.Chance(p.PPM) {
		return out
	}
	i := r.Intn(len(out))
	op := p.Operator
	if op == 0 {
		op = 1 + r.Intn(6)
	}
	switch op {
	case 1:
		out[i] = randomInstruction(r, opcodes)
	case 2:
		if len(out) < limit {
			out = append(out, vm.Instruction{})
			copy(out[i+1:], out[i:])
			out[i] = randomInstruction(r, opcodes)
		}
	case 3:
		if len(out) > 1 {
			out = append(out[:i], out[i+1:]...)
		}
	case 4:
		n := min(1+r.Intn(4), len(out)-i, limit-len(out))
		out = append(out, out[i:i+n]...)
	case 5:
		n := min(1+r.Intn(4), len(out)-i, len(out)-1)
		out = append(out[:i], out[i+n:]...)
	case 6:
		if r.Intn(2) == 0 {
			out[i].A = r.Intn(65) - 1
		} else {
			out[i].B = r.Intn(129) - 1
		}
	}
	return out
}
func Recombine(code, donor []vm.Instruction, r *RNG) []vm.Instruction {
	out := slices.Clone(code)
	if len(out) == 0 || len(donor) == 0 {
		return out
	}
	a, b := r.Intn(len(out)), r.Intn(len(donor))
	n := min(len(out)-a, len(donor)-b)
	copy(out[a:a+n], donor[b:b+n])
	return out
}
func CopyMemory(working, initial [8]int, r *RNG, p CopyPolicy) [8]int {
	source := working
	if p.Initial {
		source = initial
	}
	result := [8]int{}
	indices := []int{}
	for i := 0; i < 8; i++ {
		if p.Layout == 0 || p.Layout == 1 && i < 4 || p.Layout == 2 && i%2 == 0 {
			result[i] = source[i]
			indices = append(indices, i)
		}
	}
	if len(indices) > 0 && r.Chance(p.PPM) {
		result[indices[r.Intn(len(indices))]] = r.Intn(257) - 128
	}
	return result
}
