package experiment

import (
	"open-end/internal/evolution"
	"open-end/internal/kernel"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func TestSymbolInterventionPreservesSourceAndEmptyControl(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.Ecology = true
	c.Environment = "coupled"
	c.Symbols = "persistent"
	c.CopyModel = "evolving"
	c.MutationPPM = 0
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	hash := kernel.Hash(w)
	a, _, ar, err := ContinueSymbols(w, "persistent", 100, 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"scrambled", "unreadable"} {
		b, _, br, err := ContinueSymbols(w, mode, 100, 20)
		if err != nil {
			t.Fatal(err)
		}
		if kernel.Hash(w) != hash || ar.InitialHash != hash || ar.SourceHash != br.SourceHash || br.InitialHash == hash {
			t.Fatal("unmatched or mutated source")
		}
		b.Config.Symbols = "persistent"
		if kernel.Hash(a) != kernel.Hash(b) {
			t.Fatal("reception changes a world without words")
		}
	}
	for _, mode := range []string{"", "coupled", "unknown"} {
		if _, _, _, err := ContinueSymbols(w, mode, 100, 20); err == nil {
			t.Fatal("invalid mode accepted")
		}
	}
	w.Config.Symbols = "scrambled"
	if _, _, _, err := ContinueSymbols(w, "persistent", 100, 20); err == nil {
		t.Fatal("nonpersistent source accepted")
	}
}

// This hand-written fixture tests causal reachability, not evolved convention.
func TestSymbolReceptionCanChangeSubsequentActions(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.Ecology = true
	c.Environment = "coupled"
	c.Symbols = "persistent"
	c.MutationPPM = 0
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	p := w.Particles[1]
	p.Code = []vm.Instruction{{Op: vm.LISTEN, A: -1, B: 0}, {Op: vm.COMPARE, A: 0, B: 3}, {Op: vm.JUMP, A: 4, B: 1}, {Op: vm.TOKEN, A: -1, B: 0}, {Op: vm.ABSORB, A: 32}, {Op: vm.JUMP, A: 0}}
	w.RegisterGenome(p, "")
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
	w.Cells[p.Position].Word = &world.Word{Tokens: []int{2}, Authors: []uint64{p.ID}, Expires: 64}
	p.Energy -= 2
	w.Accounting.Dissipated += 2
	w.Symbols.Writes = 1
	w.SymbolActor(p.Genome).Writes = 1
	hash := kernel.Hash(w)
	a, _, ar, err := ContinueSymbols(w, "persistent", 4, 1)
	if err != nil {
		t.Fatal(err)
	}
	b, _, br, err := ContinueSymbols(w, "unreadable", 4, 1)
	if err != nil {
		t.Fatal(err)
	}
	if kernel.Hash(w) != hash {
		t.Fatal("source mutated")
	}
	if ar.Summary.Symbols.NonemptyReads != 1 || br.Summary.Symbols.NonemptyReads != 0 || ar.Summary.Symbols.Writes != 0 || br.Summary.Symbols.Writes != 1 || a.Particles[1].Energy <= b.Particles[1].Energy {
		t.Fatal("reception did not change the conditional action")
	}
}
