package kernel

import (
	"bytes"
	"encoding/json"
	"open-end/internal/observer"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func TestEnvironmentReplayAndCompatibility(t *testing.T) {
	for _, model := range []string{"", "evolving"} {
		c := world.DefaultConfig()
		c.Width = 12
		c.Height = 12
		c.MaxEntities = 144
		c.Ecology = true
		c.Environment = "coupled"
		c.CopyModel = model
		c.MatterDiffusion = 4
		c.ChemicalDiffusion = 4
		c.MutationPPM = 100000
		w, err := world.New(c)
		if err != nil {
			t.Fatal(err)
		}
		if err := ReloadRules(w, ruleModule(t, "baseline")); err != nil {
			t.Fatal(err)
		}
		advance(t, w, 1500)
		var buf bytes.Buffer
		if err := Save(&buf, w); err != nil {
			t.Fatal(err)
		}
		var s snapshot
		if err := json.Unmarshal(buf.Bytes(), &s); err != nil || s.Format != 5 {
			t.Fatal("wrong format")
		}
		loaded, err := Load(&buf)
		if err != nil {
			t.Fatal(err)
		}
		advance(t, w, 500)
		advance(t, loaded, 500)
		if Hash(w) != Hash(loaded) {
			t.Fatal("environment replay differs")
		}
		s.Format = 4
		b, _ := json.Marshal(s)
		if _, err := Load(bytes.NewReader(b)); err == nil {
			t.Fatal("downgraded environment snapshot accepted")
		}
		loaded.Environment.Emitted++
		if err := loaded.Validate(); err == nil {
			t.Fatal("corrupt environment ledger accepted")
		}
	}
	legacy := small(t, 1)
	legacy.Cells[0].Terrain = 1
	if err := legacy.Validate(); err == nil {
		t.Fatal("legacy terrain accepted")
	}
	legacy.Cells[0].Terrain = 0
	legacy.Particles[1].Code[0].Op = vm.BUILD
	if err := legacy.Validate(); err == nil {
		t.Fatal("new opcode in legacy world accepted")
	}
}

func TestEnvironmentWithoutMutationPreservesPhysicalControl(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 12
	c.Height = 12
	c.MaxEntities = 144
	c.Ecology = true
	c.CopyModel = "evolving"
	c.MutationPPM = 0
	c.MatterDiffusion = 4
	c.ChemicalDiffusion = 4
	legacy, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	c.Environment = "coupled"
	modern, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	advance(t, legacy, 1000)
	advance(t, modern, 1000)
	if modern.Environment.Built != 0 || modern.Environment.Emitted != 0 || observer.Observe(modern).Genomes != 1 {
		t.Fatal("engineering was seeded")
	}
	modern.Config.Environment = ""
	modern.Environment = nil
	if Hash(legacy) != Hash(modern) {
		t.Fatal("zero-mutation environment changed physics")
	}
}
