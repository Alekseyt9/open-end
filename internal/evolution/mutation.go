package evolution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"open-end/internal/vm"
)

// RNG uses SplitMix64 with explicit serialized state, independent of Go's RNG.
type RNG struct {
	State uint64 `json:"state"`
}

func (r *RNG) Uint64() uint64 {
	r.State += 0x9e3779b97f4a7c15
	z := r.State
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}
func (r *RNG) Intn(n int) int      { return int(r.Uint64() % uint64(n)) }
func (r *RNG) Chance(ppm int) bool { return ppm > 0 && r.Intn(1_000_000) < ppm }

func Hash(code []vm.Instruction, memory [8]int) string {
	b, _ := json.Marshal(struct {
		Code   []vm.Instruction
		Memory [8]int
	}{code, memory})
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func randomInstruction(r *RNG, count int) vm.Instruction {
	i := vm.Instruction{Op: vm.Opcode(r.Intn(count)), A: r.Intn(65) - 1, B: r.Intn(129) - 1}
	if i.Op == vm.CONVERT {
		i.A = r.Intn(2)
		i.B = 1 + r.Intn(8)
	}
	return i
}

// Mutate always copies its input. At most one structural mutation per copy,
// with a hard program-size bound. Memory mutation is applied during COPYMEM.
func Mutate(code []vm.Instruction, r *RNG, ppm, limit int) []vm.Instruction {
	return MutateWithOpcodes(code, r, ppm, limit, int(vm.OpcodeCount))
}

func MutateWithOpcodes(code []vm.Instruction, r *RNG, ppm, limit, opcodes int) []vm.Instruction {
	out := append([]vm.Instruction(nil), code...)
	if len(out) == 0 || !r.Chance(ppm) {
		return out
	}
	i := r.Intn(len(out))
	switch r.Intn(5) {
	case 0:
		out[i] = randomInstruction(r, opcodes)
	case 1:
		if len(out) < limit {
			out = append(out, vm.Instruction{})
			copy(out[i+1:], out[i:])
			out[i] = randomInstruction(r, opcodes)
		}
	case 2:
		if len(out) > 1 {
			out = append(out[:i], out[i+1:]...)
		}
	case 3:
		n := min(1+r.Intn(4), len(out)-i, limit-len(out))
		out = append(out, out[i:i+n]...)
	case 4:
		n := min(1+r.Intn(4), len(out)-i, len(out)-1)
		out = append(out[:i], out[i+n:]...)
	}
	return out
}
