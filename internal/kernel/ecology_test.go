package kernel

import (
	"bytes"
	"open-end/internal/observer"
	"open-end/internal/world"
	"testing"
)

func TestEcologyReplayAndConservation(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 16
	c.Height = 16
	c.MaxEntities = 256
	c.Ecology = true
	c.MatterDiffusion = 4
	c.ChemicalDiffusion = 4
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	advance(t, w, 800)
	var snapshot bytes.Buffer
	if err := Save(&snapshot, w); err != nil {
		t.Fatal(err)
	}
	clone, err := Load(&snapshot)
	if err != nil {
		t.Fatal(err)
	}
	advance(t, w, 600)
	advance(t, clone, 600)
	if Hash(w) != Hash(clone) {
		t.Fatal("ecology replay diverged")
	}
	m := observer.Observe(w)
	if m.Converted[0] == 0 || m.Converted[1] == 0 || m.Copies < 20 {
		t.Fatal("chemical seed did not replicate")
	}
}

func TestMatterTransportPreservesReplicationWithoutMutation(t *testing.T) {
	w := small(t, 1)
	w.Config.MatterDiffusion = 4
	w.Config.MutationPPM = 0
	// A sustained late copy rate matters more than surviving immortal programs.
	for i := 0; i < 20000; i++ {
		Step(w)
	}
	before := w.Accounting.Copies
	for i := 0; i < 5000; i++ {
		Step(w)
	}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
	if w.Accounting.Copies-before < 100 || len(w.Particles) < 10 {
		t.Fatal("transport baseline lost replication")
	}
}
