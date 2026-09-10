package world

import (
	"fmt"
	"open-end/internal/vm"
)

const WordLifetime uint64 = 64

// Word is a bounded physical inscription. Authors are observation provenance;
// no instruction can read author IDs. Inscription costs dissipate immediately.
type Word struct {
	Tokens  []int    `json:"tokens"`
	Authors []uint64 `json:"authors"`
	Expires uint64   `json:"expires"`
}

type SymbolCounts struct {
	Writes         uint64 `json:"writes"`
	Reads          uint64 `json:"reads"`
	NonemptyReads  uint64 `json:"nonempty_reads"`
	PairReads      uint64 `json:"pair_reads"`
	ForeignReads   uint64 `json:"foreign_reads"`
	Lookups        uint64 `json:"lookups"`
	ContextLookups uint64 `json:"context_lookups"`
}
type SymbolState struct {
	SymbolCounts
	Actors map[string]*SymbolCounts `json:"actors,omitempty"`
}

func (c Config) AllowsOpcode(op vm.Opcode) bool {
	return op < vm.OpcodeCount && (c.Symbols != "" || op < vm.EngineeringOpcodeCount) && (c.Environment != "" || op < vm.EcologyOpcodeCount)
}

func (a SymbolCounts) Values() [7]uint64 {
	return [7]uint64{a.Writes, a.Reads, a.NonemptyReads, a.PairReads, a.ForeignReads, a.Lookups, a.ContextLookups}
}
func SymbolCountsFrom(v [7]uint64) SymbolCounts {
	return SymbolCounts{v[0], v[1], v[2], v[3], v[4], v[5], v[6]}
}
func (a SymbolCounts) Sub(b SymbolCounts) (SymbolCounts, error) {
	x, y := a.Values(), b.Values()
	for i := range x {
		if x[i] < y[i] {
			return SymbolCounts{}, fmt.Errorf("symbol counter regressed")
		}
		x[i] -= y[i]
	}
	return SymbolCountsFrom(x), nil
}
func (a SymbolCounts) Validate() error {
	if a.NonemptyReads > a.Reads || a.PairReads > a.NonemptyReads || a.ForeignReads > a.NonemptyReads || a.ContextLookups > a.Lookups {
		return fmt.Errorf("invalid symbol counters")
	}
	return nil
}
func (w *World) SymbolActor(hash string) *SymbolCounts {
	if w.Symbols.Actors == nil {
		w.Symbols.Actors = map[string]*SymbolCounts{}
	}
	if w.Symbols.Actors[hash] == nil {
		w.Symbols.Actors[hash] = &SymbolCounts{}
	}
	return w.Symbols.Actors[hash]
}
func (w *World) ValidateSymbols() error {
	if (w.Config.Symbols != "") != (w.Symbols != nil) {
		return fmt.Errorf("symbol state/config mismatch")
	}
	for _, c := range w.Cells {
		if c.Word == nil {
			continue
		}
		x := c.Word
		// At the tick boundary an inscription may expire on the next inflow.
		if w.Symbols == nil || len(x.Tokens) < 1 || len(x.Tokens) > 2 || len(x.Authors) != len(x.Tokens) || x.Expires < w.Tick || x.Expires-w.Tick > WordLifetime {
			return fmt.Errorf("invalid word lifetime or storage")
		}
		for j, t := range x.Tokens {
			if t < 0 || t >= 4 || x.Authors[j] == 0 || x.Authors[j] >= w.NextID {
				return fmt.Errorf("invalid word token or author")
			}
		}
	}
	if w.Symbols == nil {
		return nil
	}
	if err := w.Symbols.SymbolCounts.Validate(); err != nil {
		return err
	}
	remaining := w.Symbols.SymbolCounts
	for hash, a := range w.Symbols.Actors {
		if a == nil || w.Genomes[hash] == nil {
			return fmt.Errorf("invalid symbol actor")
		}
		if err := a.Validate(); err != nil {
			return err
		}
		var err error
		remaining, err = remaining.Sub(*a)
		if err != nil {
			return err
		}
	}
	if remaining != (SymbolCounts{}) {
		return fmt.Errorf("symbol actors do not reconcile")
	}
	return nil
}
