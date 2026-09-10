package kernel

import (
	"bytes"
	"encoding/json"
	"open-end/internal/evolution"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func TestSymbolSnapshotReplayAndValidation(t *testing.T) {
	for _, mode := range []string{"persistent", "scrambled", "unreadable"} {
		c := world.DefaultConfig()
		c.Width = 8
		c.Height = 8
		c.MaxEntities = 64
		c.Ecology = true
		c.Environment = "coupled"
		c.Symbols = mode
		c.CopyModel = "evolving"
		w, err := world.New(c)
		if err != nil {
			t.Fatal(err)
		}
		p := w.Particles[1]
		p.Code = []vm.Instruction{{Op: vm.ABSORB, A: 32}, {Op: vm.TOKEN, A: -1, B: 0}, {Op: vm.LISTEN, A: -1, B: 1}, {Op: vm.LOOKUP, A: 1, B: 2}}
		w.RegisterGenome(p, "")
		p.Origin = evolution.Hash(p.Code, p.InitialMemory)
		w.Origins[p.Origin] = &world.Origin{Hash: p.Origin, Copies: 1}
		advance(t, w, 51)
		var buf bytes.Buffer
		if err := Save(&buf, w); err != nil {
			t.Fatal(err)
		}
		var envelope snapshot
		if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil || envelope.Format != 7 {
			t.Fatal("wrong format")
		}
		loaded, err := Load(&buf)
		if err != nil {
			t.Fatal(err)
		}
		advance(t, w, 100)
		advance(t, loaded, 100)
		if Hash(w) != Hash(loaded) {
			t.Fatal("symbol replay mismatch")
		}
		for f := 2; f <= 6; f++ {
			envelope.Format = f
			b, _ := json.Marshal(envelope)
			if _, err := Load(bytes.NewReader(b)); err == nil {
				t.Fatal("downgrade accepted")
			}
		}
		w.Symbols.Writes++
		if err := w.Validate(); err == nil {
			t.Fatal("bad ledger accepted")
		}
		w.Symbols.Writes--
		word := w.Cells[p.Position].Word
		word.Tokens[0] = 4
		if err := w.Validate(); err == nil {
			t.Fatal("invalid alphabet accepted")
		}
		word.Tokens[0] = 0
		word.Expires = w.Tick + 65
		if err := w.Validate(); err == nil {
			t.Fatal("invalid expiry accepted")
		}
	}
}
func TestSymbolZeroMutationPreservesEngineeringWorld(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.Ecology = true
	c.Environment = "coupled"
	c.CopyModel = "evolving"
	c.MutationPPM = 0
	legacy, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	c.Symbols = "persistent"
	modern, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	advance(t, legacy, 1000)
	advance(t, modern, 1000)
	if modern.Symbols.SymbolCounts != (world.SymbolCounts{}) {
		t.Fatal("symbol behavior was seeded")
	}
	modern.Symbols = nil
	modern.Config.Symbols = ""
	if Hash(legacy) != Hash(modern) {
		t.Fatal("unexercised symbolic physics changes legacy world")
	}
	if legacy.Config.AllowsOpcode(vm.TOKEN) || !legacy.Config.AllowsOpcode(vm.EMIT) {
		t.Fatal("legacy repertoire changed")
	}
}
