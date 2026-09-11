package kernel

import (
	"bytes"
	"encoding/json"
	"open-end/internal/dsl"
	"open-end/internal/world"
	"testing"
)

func TestMetabolicSwitchSnapshotReplayAndDowngradeRejection(t *testing.T) {
	for _, motion := range []string{"", "yielding"} {
		c := world.DefaultConfig()
		c.Width = 8
		c.Height = 8
		c.MaxEntities = 64
		c.Ecology = true
		c.Environment = "coupled"
		c.Symbols = "persistent"
		c.CopyModel = "evolving"
		c.BondMotion = motion
		c.MetabolicSwitchCost = 2
		w, err := world.New(c)
		if err != nil {
			t.Fatal(err)
		}
		advance(t, w, 1)
		if w.Particles[1] == nil || w.Particles[1].MetabolicState != 1 {
			t.Fatal("prepared state not present at snapshot boundary")
		}
		var b bytes.Buffer
		if err := Save(&b, w); err != nil {
			t.Fatal(err)
		}
		var s snapshot
		json.Unmarshal(b.Bytes(), &s)
		if s.Format != 9 {
			t.Fatal("wrong format")
		}
		q, err := Load(bytes.NewReader(b.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		advance(t, w, 300)
		advance(t, q, 300)
		if Hash(w) != Hash(q) {
			t.Fatal("switch replay differs")
		}
		for f := 2; f < 9; f++ {
			s.Format = f
			raw, _ := json.Marshal(s)
			if _, err := Load(bytes.NewReader(raw)); err == nil {
				t.Fatal("downgrade accepted", f)
			}
		}
		if err := ReloadRules(w, nil); err == nil {
			t.Fatal("DSL admitted")
		}
		if err := ScheduleRules(w, dsl.Change{}); err == nil {
			t.Fatal("DSL schedule admitted")
		}
		for _, p := range s.World.Particles {
			p.MetabolicState = 3
			break
		}
		if s.World.Validate() == nil {
			t.Fatal("invalid preparation accepted")
		}
	}
}
