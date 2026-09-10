package observer

import (
	"open-end/internal/world"
	"testing"
)

func TestDescendantBranchIncludesActualSingleFounderAncestry(t *testing.T) {
	w, tr, p := collectiveFixture(t)
	groupAdvance(w, tr, 2)
	groupCopy(w, tr, p[0], p[2])
	groupCopy(w, tr, p[2], p[3])
	if len(p[3].Code) == 0 {
		t.Fatal("fixture did not copy along adjacent path")
	}
	w.Relations[world.RelationKey(3, 4)] = world.Relation{A: 3, B: 4}
	groupAdvance(w, tr, 3)
	bs := tr.DescendantBranches()
	if len(bs) != 1 || bs[0].Contributors != 1 || !bs[0].ParentPresent || len(bs[0].Members) != 2 {
		t.Fatalf("clonal branch missed: %+v", bs)
	}
	if len(tr.Frame(w).Telemetry.Collectives.Candidates) != 0 {
		t.Fatal("old multi-founder criterion changed")
	}
	bs[0].Members[0] = 999
	if tr.DescendantBranches()[0].Members[0] == 999 {
		t.Fatal("returned members alias observer")
	}
	delete(w.Relations, world.RelationKey(1, 2))
	groupAdvance(w, tr, 1)
	if tr.DescendantBranches()[0].ParentPresent {
		t.Fatal("missing parent not detected")
	}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
}
