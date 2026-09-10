package kernel

import (
	"bytes"
	"encoding/json"
	"open-end/internal/world"
	"testing"
)

func TestBondMotionSnapshotVersionAndReplay(t *testing.T) {
	for _, symbols := range []string{"", "persistent"} {
		c := world.DefaultConfig()
		c.Width = 8
		c.Height = 8
		c.MaxEntities = 64
		c.Ecology = true
		c.Environment = "coupled"
		c.CopyModel = "evolving"
		c.Symbols = symbols
		c.BondMotion = "yielding"
		c.MutationPPM = 100000
		w, err := world.New(c)
		if err != nil {
			t.Fatal(err)
		}
		advance(t, w, 700)
		var buf bytes.Buffer
		if err := Save(&buf, w); err != nil {
			t.Fatal(err)
		}
		var s snapshot
		if err := json.Unmarshal(buf.Bytes(), &s); err != nil || s.Format != 8 {
			t.Fatal("wrong snapshot version")
		}
		loaded, err := Load(&buf)
		if err != nil {
			t.Fatal(err)
		}
		advance(t, w, 300)
		advance(t, loaded, 300)
		if Hash(w) != Hash(loaded) {
			t.Fatal("yielding replay differs")
		}
		for format := 2; format <= 7; format++ {
			s.Format = format
			b, _ := json.Marshal(s)
			if _, err := Load(bytes.NewReader(b)); err == nil {
				t.Fatal("downgrade accepted")
			}
		}
	}
}
