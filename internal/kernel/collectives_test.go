package kernel

import (
	"bytes"
	"encoding/json"
	"open-end/internal/observer"
	"open-end/internal/world"
	"testing"
)

func TestCollectiveObservationAndAblationReplay(t *testing.T) {
	for _, mode := range []string{"", "bonds", "sharing", "signal-reading"} {
		c := world.DefaultConfig()
		c.Width, c.Height, c.MaxEntities = 12, 12, 144
		c.Ecology, c.Environment, c.CopyModel = true, "coupled", "evolving"
		c.MatterDiffusion, c.ChemicalDiffusion, c.MutationPPM = 4, 4, 100000
		c.CollectiveAblation = mode
		w, err := world.New(c)
		if err != nil {
			t.Fatal(err)
		}
		plain, err := world.New(c)
		if err != nil {
			t.Fatal(err)
		}
		tracker := observer.NewTracker(w)
		if err := tracker.EnableCollectives(w, w, 10); err != nil {
			t.Fatal(err)
		}
		frames := []observer.Metrics{tracker.Frame(w)}
		for i := 0; i < 1500; i++ {
			StepObserved(w, tracker)
			Step(plain)
			if (i+1)%100 == 0 {
				frames = append(frames, tracker.Frame(w))
			}
		}
		if Hash(w) != Hash(plain) {
			t.Fatal("observer changed physics")
		}
		if _, err := observer.Summarize(frames, 1500); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := Save(&buf, w); err != nil {
			t.Fatal(err)
		}
		var s snapshot
		if err := json.Unmarshal(buf.Bytes(), &s); err != nil {
			t.Fatal(err)
		}
		want := 5
		if mode != "" {
			want = 6
		}
		if s.Format != want {
			t.Fatal("wrong snapshot format")
		}
		loaded, err := Load(&buf)
		if err != nil {
			t.Fatal(err)
		}
		advance(t, w, 500)
		advance(t, loaded, 500)
		if Hash(w) != Hash(loaded) {
			t.Fatal("ablation replay differs")
		}
		if mode != "" {
			s.Format = 5
		} else {
			s.Format = 6
		}
		b, _ := json.Marshal(s)
		if _, err := Load(bytes.NewReader(b)); err == nil {
			t.Fatal("incompatible format accepted")
		}
	}
}
