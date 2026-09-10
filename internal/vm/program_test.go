package vm

import "testing"

func TestJumpAndBounds(t *testing.T) {
	code := []Instruction{{Op: JUMP, A: -1, B: 1}, {Op: NOP}, {Op: NOP}}
	if Decode(code, 0, true).NextIP != 2 || Decode(code, 0, false).NextIP != 1 {
		t.Fatal("conditional jump")
	}
	if Decode(code, -3, true).NextIP != 2 {
		t.Fatal("negative address")
	}
	if Decode(nil, 0, false).NextIP != 0 {
		t.Fatal("empty program")
	}
}
