// Package vm defines bounded executable matter. Instructions only produce intents;
// the kernel resolves their physical effects.
package vm

type Opcode uint8

const (
	NOP Opcode = iota
	SENSE
	MOVE
	ABSORB
	ALLOCATE
	COPY
	COPYMEM
	TRANSFER
	WRITE
	READ
	COMPARE
	JUMP
	CONVERT
	TARGET
	TAKE
	BIND
	UNBIND
	BUILD
	EMIT
	TOKEN
	LISTEN
	LOOKUP
	OpcodeCount
)

const BaselineOpcodeCount = CONVERT
const EcologyOpcodeCount = BUILD
const EngineeringOpcodeCount = TOKEN

type Instruction struct {
	Op Opcode `json:"op"`
	A  int    `json:"a"`
	B  int    `json:"b"`
}

type Intent struct {
	Instruction
	NextIP int
}

// Decode has no access to mutable world state. JUMP B=0 is unconditional;
// B=1 tests the comparison flag, B=-1 tests its inverse.
func Decode(code []Instruction, ip int, flag bool) Intent {
	if len(code) == 0 {
		return Intent{}
	}
	ip = Index(ip, len(code))
	i := code[ip]
	next := (ip + 1) % len(code)
	if i.Op == JUMP && (i.B == 0 || i.B == 1 && flag || i.B == -1 && !flag) {
		next = Index(i.A, len(code))
	}
	return Intent{Instruction: i, NextIP: next}
}

func Index(n, size int) int { return ((n % size) + size) % size }

// Seed is a program, not a privileged reproduction rule. Mutants execute the
// same instructions under the same resource constraints.
func Seed() []Instruction {
	return []Instruction{
		{Op: ABSORB, A: 32},
		{Op: SENSE, A: 0, B: 0}, // own energy -> memory[0]
		{Op: COMPARE, A: 0, B: 100},
		{Op: JUMP, A: 6, B: 1},
		{Op: MOVE, A: -1},
		{Op: JUMP, A: 0},
		{Op: ALLOCATE, A: -1},
		{Op: COPYMEM},
		{Op: COPY},
		{Op: TRANSFER, A: 48},
		{Op: MOVE, A: -1},
		{Op: JUMP, A: 0},
	}
}

// EcologySeed is one generalist seed program. Both reactions are initially
// available to the same executable matter; no specialist populations are seeded.
func EcologySeed() []Instruction {
	code := []Instruction{{Op: CONVERT, A: 0, B: 8}, {Op: CONVERT, A: 1, B: 8}}
	code = append(code, Seed()[1:]...)
	code[4].A = 7 // branch past the extra instruction to ALLOCATE
	return code
}
