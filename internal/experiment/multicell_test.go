package experiment

import (
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
	"testing"
)

func TestClonalMaturityRequiresContinuousParentAndDeduplicates(t *testing.T) {
	s := &multicellSink{minAge: 2, pending: map[string]uint64{}, credited: map[string]int{}}
	b := observer.DescendantBranch{Cohort: 1, Members: []uint64{3, 4}, Contributors: 1, ParentPresent: true, FounderIDs: []uint64{1}}
	s.recordBranches(1, []observer.DescendantBranch{b})
	b.ParentPresent = false
	s.recordBranches(2, []observer.DescendantBranch{b})
	b.ParentPresent = true
	s.recordBranches(3, []observer.DescendantBranch{b})
	s.recordBranches(4, []observer.DescendantBranch{b})
	if len(s.d.Clonal) != 0 {
		t.Fatal("disjoint episodes accumulated")
	}
	s.recordBranches(5, []observer.DescendantBranch{b})
	if len(s.d.Clonal) != 1 || s.d.Clonal[0].Since != 3 {
		t.Fatal("mature clonal branch missed")
	}
	s.recordBranches(6, nil)
	for tick := uint64(7); tick <= 9; tick++ {
		s.recordBranches(tick, []observer.DescendantBranch{b})
	}
	if len(s.d.Clonal) != 1 || s.d.Clonal[0].Last != 9 {
		t.Fatal("rejoined branch counted twice")
	}
	b.Contributors = 2
	for tick := uint64(10); tick <= 13; tick++ {
		s.recordBranches(tick, []observer.DescendantBranch{b})
	}
	if len(s.d.Clonal) != 1 {
		t.Fatal("multi-founder branch credited as clonal")
	}
}
func TestMulticellControlMatchesOriginalGroupAssay(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.Ecology = true
	c.Environment = "coupled"
	c.CopyModel = "evolving"
	c.MutationPPM = 100000
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	hash := kernel.Hash(w)
	a, _, ar, err := ContinueMulticell(w, "intact", 300, 100, 10)
	if err != nil {
		t.Fatal(err)
	}
	b, _, _, err := ContinueCollectives(w, "intact", 300, 100, 10)
	if err != nil {
		t.Fatal(err)
	}
	if kernel.Hash(a) != kernel.Hash(b) || kernel.Hash(w) != hash || ar.InitialHash != hash {
		t.Fatal("diagnostics changed source or control")
	}
	_, _, yr, err := ContinueMulticell(w, "yielding", 300, 100, 10)
	if err != nil {
		t.Fatal(err)
	}
	if yr.SourceHash != hash || yr.InitialHash == hash || kernel.Hash(w) != hash {
		t.Fatal("unmatched treatment")
	}
}
