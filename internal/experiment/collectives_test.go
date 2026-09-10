package experiment

import (
	"open-end/internal/kernel"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func TestCollectiveAssayPreservesSourceAndInactiveControls(t *testing.T) {
	c := world.DefaultConfig()
	c.Width, c.Height, c.MaxEntities = 8, 8, 64
	c.Ecology, c.Environment, c.CopyModel = true, "coupled", "evolving"
	c.MutationPPM = 0
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	hash := kernel.Hash(w)
	var intactHash, reference string
	for _, mode := range []string{"intact", "bonds", "sharing", "signal-reading"} {
		end, frames, r, err := ContinueCollectives(w, mode, 100, 20, 10)
		if err != nil {
			t.Fatal(err)
		}
		if kernel.Hash(w) != hash || r.SourceHash != hash || r.Summary.Collectives == nil {
			t.Fatal("source mutation or missing evidence")
		}
		if mode == "intact" {
			reference = r.Summary.Collectives.End.ReferenceHash
			intactHash = kernel.Hash(end)
		}
		if r.Summary.Collectives.End.ReferenceHash != reference || len(frames) != 6 {
			t.Fatal("unmatched observation")
		}
		end.Config.CollectiveAblation = ""
		if kernel.Hash(end) != intactHash {
			t.Fatal("inactive control changed physical state")
		}
	}
}

func TestBondAblationKeepsUnmodifiedReferenceGroups(t *testing.T) {
	c := world.DefaultConfig()
	c.Width, c.Height, c.MaxEntities = 8, 8, 64
	c.Ecology, c.Environment, c.MutationPPM = true, "coupled", 0
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, instruction := range []vm.Instruction{{Op: vm.ALLOCATE, A: 1}, {Op: vm.COPY}, {Op: vm.BIND}} {
		rules.Resolve(w, rules.Event{Actor: 1, Intent: vm.Intent{Instruction: instruction}})
	}
	if len(w.Relations) != 1 {
		t.Fatal("fixture has no bond")
	}
	hash := kernel.Hash(w)
	_, frames, _, err := ContinueCollectives(w, "bonds", 10, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	cg := frames[0].Telemetry.Collectives
	if len(cg.Reference) != 1 || len(cg.Reference[0].Members) != 2 || len(cg.Current) != 0 || cg.FoundersAlive != 2 || cg.Intact != 0 {
		t.Fatal("ablation erased the shared reference")
	}
	if kernel.Hash(w) != hash {
		t.Fatal("ablation mutated source")
	}
}
