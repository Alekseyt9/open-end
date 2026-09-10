package experiment

import (
	"open-end/internal/kernel"
	"open-end/internal/world"
	"testing"
)

func TestEnvironmentInterventionPreservesSourceAndControls(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.Ecology = true
	c.Environment = "coupled"
	c.CopyModel = "evolving"
	c.MutationPPM = 0
	c.MatterDiffusion = 4
	c.ChemicalDiffusion = 4
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	hash := kernel.Hash(w)
	a, _, ar, err := ContinueEnvironment(w, "coupled", 100, 20)
	if err != nil {
		t.Fatal(err)
	}
	b, _, br, err := ContinueEnvironment(w, "inert", 100, 20)
	if err != nil {
		t.Fatal(err)
	}
	if hash != kernel.Hash(w) || ar.InitialHash != hash || ar.SourceHash != br.SourceHash {
		t.Fatal("source mutated or unmatched")
	}
	b.Config.Environment = "coupled"
	if kernel.Hash(a) != kernel.Hash(b) {
		t.Fatal("empty-field control differs")
	}
	// Construct a physical field in the fixture, transferring existing matter.
	w.Tick = 1
	pos := w.Neighbor(w.Particles[1].Position, 1)
	w.Cells[pos].Matter--
	w.Cells[pos].Terrain++
	w.Environment.Built = 1
	w.EnvironmentActor(w.Particles[1].Genome).Built = 1
	_, _, ar, err = ContinueEnvironment(w, "coupled", 20, 5)
	if err != nil {
		t.Fatal(err)
	}
	_, _, br, err = ContinueEnvironment(w, "inert", 20, 5)
	if err != nil {
		t.Fatal(err)
	}
	if ar.Summary.Environment.BlockedLight == 0 || br.Summary.Environment.BlockedLight != 0 {
		t.Fatal("intervention failed to isolate shading")
	}
}
