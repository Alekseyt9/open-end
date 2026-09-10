package kernel

import (
	"bytes"
	"encoding/json"
	"open-end/internal/evolution"
	"open-end/internal/observer"
	"open-end/internal/world"
	"testing"
)

func TestEncodedCopyReplayAndControls(t *testing.T) {
	for _, model := range []string{"fixed", "evolving"} {
		for _, dsl := range []bool{false, true} {
			c := world.DefaultConfig()
			c.Width = 12
			c.Height = 12
			c.MaxEntities = 144
			c.Ecology = true
			c.MatterDiffusion = 4
			c.ChemicalDiffusion = 4
			c.CopyModel = model
			w, err := world.New(c)
			if err != nil {
				t.Fatal(err)
			}
			if dsl {
				if err := ReloadRules(w, ruleModule(t, "baseline")); err != nil {
					t.Fatal(err)
				}
			}
			var zero bytes.Buffer
			if err := Save(&zero, w); err != nil {
				t.Fatal(err)
			}
			loaded, err := Load(&zero)
			if err != nil || Hash(w) != Hash(loaded) {
				t.Fatalf("empty ledger round trip: %v", err)
			}
			advance(t, w, 1000)
			var buf bytes.Buffer
			if err := Save(&buf, w); err != nil {
				t.Fatal(err)
			}
			var s snapshot
			if err := json.Unmarshal(buf.Bytes(), &s); err != nil || s.Format != 4 {
				t.Fatal("wrong snapshot format")
			}
			loaded, err = Load(&buf)
			if err != nil {
				t.Fatal(err)
			}
			advance(t, w, 500)
			advance(t, loaded, 500)
			if Hash(w) != Hash(loaded) || len(w.Variation) == 0 {
				t.Fatal("replay or copy ledger failed")
			}
			if model == "fixed" {
				for _, r := range w.Variation {
					want := evolution.CopyPolicy{Kind: r.Policy.Kind, PPM: c.MutationPPM}
					if r.Policy != want {
						t.Fatal("fixed operands affected policy")
					}
				}
			}
			s.Format = 2
			b, _ := json.Marshal(s)
			if _, err := Load(bytes.NewReader(b)); err == nil {
				t.Fatal("downgraded snapshot accepted")
			}
			for _, r := range loaded.Variation {
				if r.Policy.Kind != "code" {
					continue
				}
				r.Copies++
				break
			}
			if err := loaded.Validate(); err == nil {
				t.Fatal("corrupt copy ledger accepted")
			}
		}
	}
	c := world.DefaultConfig()
	c.Width = 12
	c.Height = 12
	c.MaxEntities = 144
	c.CopyModel = "evolving"
	c.MutationPPM = 0
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	advance(t, w, 1000)
	if m := observer.Observe(w); m.Genomes != 1 || m.Copies == 0 {
		t.Fatal("control did not replicate faithfully")
	}
	for _, r := range w.Variation {
		if r.Changed != 0 || r.Recombined != 0 {
			t.Fatal("zero-mutation changed inheritance")
		}
	}
}
