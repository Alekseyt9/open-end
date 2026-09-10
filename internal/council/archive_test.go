package council

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParetoTradeoffsAndRareDominatedCell(t *testing.T) {
	// The rare cell has weak quality and nearby historical behavior. Its
	// novelty is lower too, so cell protection must override Pareto exclusion.
	points := []ArchivePoint{
		{Node: "a", Context: "same", Horizon: "now", Cell: "common", Descriptor: []float64{10}, Eligible: true, Tip: true, Diversity: 10, Structure: 10, Persistence: 1},
		{Node: "b", Context: "same", Horizon: "now", Cell: "rare", Descriptor: []float64{0}, Eligible: true, Tip: true, Diversity: 1, Structure: 1, Persistence: 1},
		{Node: "c", Context: "same", Horizon: "now", Cell: "rare", Descriptor: []float64{.2}, Eligible: true, Tip: true, Diversity: .5, Structure: .5, Persistence: 1},
		{Node: "d", Context: "other", Horizon: "other-time", Cell: "elsewhere", Descriptor: []float64{1}, Eligible: true, Tip: true, Diversity: 1000, Structure: 1000, Persistence: 1},
		{Node: "h", Context: "same", Horizon: "old", Cell: "rare", Descriptor: []float64{.1}, Eligible: true, Diversity: 1, Structure: 1, Persistence: 1},
	}
	got := rankArchive(points, DefaultArchiveOptions())
	if !got[0].Pareto || got[1].Pareto || !got[1].Recommended || got[2].Recommended {
		t.Fatal("rare dominated cell lost or ordinary dominated candidate retained", got)
	}
	if !reflect.DeepEqual(got[1].DominatedBy, []string{"a"}) {
		t.Fatal("missing explanation")
	}
	if got[0].Novelty == nil || *got[0].Novelty <= *got[1].Novelty {
		t.Fatal("fixture must dominate on novelty as well")
	}
	if !got[3].Pareto || got[3].Novelty != nil {
		t.Fatal("incompatible context compared")
	}
	a := ArchivePoint{Diversity: 10, Structure: 1, Persistence: 1}
	b := ArchivePoint{Diversity: 1, Structure: 10, Persistence: 1}
	if archiveDominates(a, b) || archiveDominates(b, a) || archiveDominates(a, a) {
		t.Fatal("tradeoff or equality eliminated")
	}
}

func TestNoveltyPermutationDuplicatesAndHistoricalRecovery(t *testing.T) {
	makePoints := func() []ArchivePoint {
		return []ArchivePoint{
			{Node: "a", Context: "x", Horizon: "old", Cell: "lost", Descriptor: []float64{0}, Eligible: true, Diversity: 1, Persistence: 1},
			{Node: "b", Context: "x", Horizon: "new", Cell: "current", Descriptor: []float64{2}, Eligible: true, Tip: true, Diversity: 2, Persistence: 1},
			{Node: "c", Context: "x", Horizon: "new", Cell: "current", Descriptor: []float64{2}, Eligible: true, Tip: true, Diversity: 2, Persistence: 1},
		}
	}
	got := rankArchive(makePoints(), DefaultArchiveOptions())
	if !got[0].Recommended || !got[0].Protected {
		t.Fatal("lost historical cell not preserved")
	}
	if got[0].Novelty == nil || *got[0].Novelty != 2 || len(got[0].Nearest) != 1 {
		t.Fatal("duplicates distorted reference density")
	}
	shuffled := makePoints()
	shuffled[0], shuffled[2] = shuffled[2], shuffled[0]
	if !reflect.DeepEqual(got, rankArchive(shuffled, DefaultArchiveOptions())) {
		t.Fatal("input order changes decision")
	}
}

func TestArchiveValidationPersistenceAndNoSimulationMutation(t *testing.T) {

	input, _, _, _ := fixture(t)
	dir := filepath.Join(t.TempDir(), "tree")
	if err := InitTree(input, dir, "", 200); err != nil {
		t.Fatal(err)
	}
	before, _ := ReadTree(dir)
	selection, err := os.ReadFile(filepath.Join(dir, "selection.json"))
	if err != nil {
		t.Fatal(err)
	}
	// Sparse short fixtures fail the default stable-window gate.
	a, err := ArchiveTree(dir, DefaultArchiveOptions(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Recommended) != 0 || a.Points[0].Eligible {
		t.Fatal("short history admitted")
	}
	if _, err := ArchiveTree(dir, DefaultArchiveOptions(), true); err == nil {
		t.Fatal("empty selection applied")
	}
	afterSelection, _ := os.ReadFile(filepath.Join(dir, "selection.json"))
	if string(afterSelection) != string(selection) {
		t.Fatal("empty decision changed selection")
	}
	files, _ := os.ReadDir(filepath.Join(dir, "archive"))
	if len(files) != 1 {
		t.Fatal("identical refresh created revisions")
	}
	latest, err := LatestArchive(dir)
	if err != nil || latest.ID != a.ID {
		t.Fatal("cannot read decision", err)
	}
	after, _ := ReadTree(dir)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("archive changed tree")
	}
	// Archive validation itself cannot bypass physical snapshot validation.
	path := filepath.Join(nodeRound(dir, "b000001"), "w001.snapshot.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ArchiveTree(dir, DefaultArchiveOptions(), false); err == nil {
		t.Fatal("modified snapshot accepted")
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := GrowTree(dir, nil, false, TrialOptions{Ticks: 100, Every: 20, Window: 80, Workers: 2}, io.Discard); err != nil {
		t.Fatal(err)
	}
	view, err := ReadTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if latest.TreeHash == jsonHash(view.Nodes) {
		t.Fatal("stale archive undetected")
	}
	next, err := ArchiveTree(dir, DefaultArchiveOptions(), false)
	if err != nil {
		t.Fatal(err)
	}
	if next.ID == a.ID {
		t.Fatal("new nodes did not create a revision")
	}
	old, err := readArchiveDecision(filepath.Join(dir, "archive", "a000001", "decision.json"))
	if err != nil || old.ID != a.ID {
		t.Fatal("old decision changed")
	}
	for _, bad := range []ArchiveOptions{{0, .75, .75, 1, 3}, {10000, math.NaN(), .75, 1, 3}, {10000, .75, 1.1, 1, 3}, {10000, .75, .75, 0, 3}, {10000, .75, .75, 1, 0}} {
		if _, err := ArchiveTree(dir, bad, false); err == nil {
			t.Fatal("invalid options accepted")
		}
	}
}
