package council

import (
	"io"
	"path/filepath"
	"reflect"
	"testing"
)

func TestArchiveEligibleApplyAndPhysicalParity(t *testing.T) {
	input, _, _, _ := fixture(t)
	dir := filepath.Join(t.TempDir(), "tree")
	if err := InitTree(input, dir, "", 200); err != nil {
		t.Fatal(err)
	}
	children, err := GrowTree(dir, nil, false, TrialOptions{Ticks: 10000, Every: 1000, Window: 10000, Workers: 2}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	before, err := ReadTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := SelectTree(dir, []string{"b000001"}); err != nil {
		t.Fatal(err)
	}
	a, err := ArchiveTree(dir, DefaultArchiveOptions(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.Recommended, children) || !a.Points[1].Eligible || a.Points[1].Novelty != nil {
		t.Fatal("singleton cohort should use available objectives", a)
	}
	v, err := ReadTree(dir)
	if err != nil || v.Selected[0] != "b000001" {
		t.Fatal("preview applied selection", err)
	}
	applied, err := ArchiveTree(dir, DefaultArchiveOptions(), true)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != applied.ID {
		t.Fatal("applying changed evidence")
	}
	after, err := ReadTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after.Selected, children) || !reflect.DeepEqual(after.Nodes, before.Nodes) {
		t.Fatal("selection changed world evidence")
	}
	for _, n := range before.Nodes {
		if _, err := checkRound(nodeRound(dir, n.ID), "", true); err != nil {
			t.Fatal("physical state changed", err)
		}
	}
	// Editing a published archive without updating its identity is rejected.
	path := filepath.Join(dir, "archive", "a000001", "decision.json")
	a.Recommended = []string{"b000001"}
	if err := replaceJSON(path, a); err != nil {
		t.Fatal(err)
	}
	if _, err := LatestArchive(dir); err == nil {
		t.Fatal("modified archive accepted")
	}
}

func TestParetoDoesNotCompareDifferentHorizons(t *testing.T) {
	points := []ArchivePoint{
		{Node: "a", Context: "x", Horizon: "early", Cell: "same", Descriptor: []float64{1}, Eligible: true, Tip: true, Diversity: 1, Structure: 1, Persistence: 1},
		{Node: "b", Context: "x", Horizon: "late", Cell: "same", Descriptor: []float64{1}, Eligible: true, Tip: true, Diversity: 100, Structure: 100, Persistence: 1},
	}
	got := rankArchive(points, DefaultArchiveOptions())
	if !got[0].Pareto || !got[1].Pareto {
		t.Fatal("different final ticks ranked together")
	}
}
