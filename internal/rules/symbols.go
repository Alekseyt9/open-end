package rules

import (
	"open-end/internal/vm"
	"open-end/internal/world"
)

// Scrambling changes the alphabet every tick without consuming either RNG.
// The transform preserves length/order/equality within one tick, but a stable
// sender/receiver convention can no longer rely on a fixed numeric identity.
func readWord(w *world.World, p *world.Particle, direction int) (int, int, bool) {
	x := w.Cells[engineeringCell(w, p, direction)].Word
	if x == nil || x.Expires <= w.Tick || w.Config.Symbols == "unreadable" {
		return 0, 0, false
	}
	shift := 0
	if w.Config.Symbols == "scrambled" {
		z := w.Tick + 0x9e3779b97f4a7c15
		z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
		z = (z ^ (z >> 27)) * 0x94d049bb133111eb
		shift = int((z ^ (z >> 31)) & 3)
	}
	value, foreign := 0, false
	for j, t := range x.Tokens {
		value = value*4 + vm.Index(t+shift, 4)
		foreign = foreign || x.Authors[j] != p.ID
	}
	if len(x.Tokens) == 1 {
		value++
	} else {
		value += 5
	}
	return value, len(x.Tokens), foreign
}

func applySymbol(w *world.World, p *world.Particle, e Event) {
	i := e.Intent.Instruction
	a := w.SymbolActor(p.Genome)
	counts := []*world.SymbolCounts{&w.Symbols.SymbolCounts, a}
	switch i.Op {
	case vm.TOKEN:
		c := &w.Cells[engineeringCell(w, p, i.A)]
		if c.Word == nil || c.Word.Expires <= w.Tick {
			c.Word = &world.Word{}
		}
		x := c.Word
		if len(x.Tokens) == 2 {
			x.Tokens = x.Tokens[1:]
			x.Authors = x.Authors[1:]
		}
		x.Tokens = append(x.Tokens, vm.Index(p.Memory[vm.Index(i.B, 8)], 4))
		x.Authors = append(x.Authors, p.ID)
		x.Expires = w.Tick + world.WordLifetime
		for _, c := range counts {
			c.Writes++
		}
	case vm.LISTEN:
		p.Memory[vm.Index(i.B, 8)] = e.Sensed
		for _, c := range counts {
			c.Reads++
			if e.WordLength > 0 {
				c.NonemptyReads++
			}
			if e.WordLength == 2 {
				c.PairReads++
			}
			if e.ForeignWord {
				c.ForeignReads++
			}
		}
	case vm.LOOKUP:
		input, output := vm.Index(i.A, 8), vm.Index(i.B, 8)
		context := vm.Index(p.Memory[output], 8)
		// Reduce each operand before addition to avoid signed integer overflow.
		index := (vm.Index(p.Memory[input], 8) + context) % 8
		value := p.Memory[index]
		p.Memory[output] = value
		for _, c := range counts {
			c.Lookups++
			if context != 0 {
				c.ContextLookups++
			}
		}
	}
}
