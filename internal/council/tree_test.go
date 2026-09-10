package council

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"open-end/internal/kernel"
)

func TestTreeMultipleGenerationsAndReplay(t *testing.T) {
	input, _, _, response := fixture(t)
	trees := []string{}
	for _, workers := range []int{1, 16} {
		dir := filepath.Join(t.TempDir(), "tree")
		if err := InitTree(input, dir, "", 200); err != nil {
			t.Fatal(err)
		}
		v, err := ReadTree(dir)
		if err != nil {
			t.Fatal(err)
		}
		root := v.Nodes[0]
		response.RequestID = root.RequestID
		save := filepath.Join(nodeRound(dir, root.ID), "response.json")
		if err := writeJSON(save, response); err != nil {
			t.Fatal(err)
		}
		opts := TrialOptions{Ticks: 100, Every: 20, Window: 80, Workers: workers}
		children, err := GrowTree(dir, nil, true, opts, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		if len(children) != 2 {
			t.Fatalf("want two variants: %v", children)
		}
		v, err = ReadTree(dir)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(v.Selected, children) {
			t.Fatal("children not retained")
		}
		if len(v.Nodes) != 3 || v.Runs[0].Status != "complete" {
			t.Fatal("incomplete first generation")
		}
		// No new AI response is needed to retain and advance both directions.
		grandchildren, err := GrowTree(dir, nil, false, opts, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		if len(grandchildren) != 2 {
			t.Fatal("lost a retained direction")
		}
		v, err = ReadTree(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(v.Nodes) != 5 || v.Nodes[3].Generation != 2 || v.Nodes[4].Generation != 2 {
			t.Fatal("broken ancestry")
		}
		if v.Nodes[3].Parent != children[0] || v.Nodes[4].Parent != children[1] {
			t.Fatal("wrong parent")
		}
		// The control-control path must match a direct 200-tick continuation.
		for i, wb := range v.Nodes[0].Worlds {
			w, err := loadWorld(filepath.Join(nodeRound(dir, root.ID), wb.ID+".snapshot.json"))
			if err != nil {
				t.Fatal(err)
			}
			for tick := 0; tick < 200; tick++ {
				kernel.Step(w)
			}
			if kernel.Hash(w) != v.Nodes[3].Worlds[i].SnapshotHash {
				t.Fatal("tree changes control physics")
			}
			if v.Nodes[0].Worlds[i].SnapshotHash != wb.SnapshotHash {
				t.Fatal("root mutated")
			}
		}
		// Deselecting branches never deletes their snapshots or ancestry.
		if err := SelectTree(dir, []string{grandchildren[0]}); err != nil {
			t.Fatal(err)
		}
		v, err = ReadTree(dir)
		if err != nil || len(v.Nodes) != 5 || len(v.Selected) != 1 {
			t.Fatal("selection is destructive", err)
		}
		report := filepath.Join(dir, "index.html")
		if err := ExportTree(dir, report); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(report)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "const model =") || strings.Contains(string(data), "ZgotmplZ") {
			t.Fatal("bad report")
		}
		if err := ExportTree(dir, report); err != nil {
			t.Fatal("cannot refresh report", err)
		}
		trees = append(trees, dir)
	}
	a, _ := ReadTree(trees[0])
	b, _ := ReadTree(trees[1])
	for i := range a.Nodes {
		if !reflect.DeepEqual(a.Nodes[i].Worlds, b.Nodes[i].Worlds) {
			t.Fatal("worker count changes tree results")
		}
	}
}

func TestTreeRejectsInvalidInputsAndRetainsFailedRun(t *testing.T) {
	input, _, _, response := fixture(t)
	dir := filepath.Join(t.TempDir(), "tree")
	if err := InitTree(input, dir, "", 200); err != nil {
		t.Fatal(err)
	}
	if err := InitTree(input, dir, "", 200); err == nil {
		t.Fatal("overwrote tree")
	}
	for _, ids := range [][]string{nil, {"../bad"}, {"b000001", "b000001"}} {
		if err := SelectTree(dir, ids); err == nil {
			t.Fatal("invalid selection accepted", ids)
		}
	}
	opts := TrialOptions{Ticks: 100, Every: 20, Window: 80, Workers: 2}
	if _, err := GrowTree(dir, nil, true, opts, io.Discard); err == nil {
		t.Fatal("missing proposal accepted")
	}
	v, _ := ReadTree(dir)
	if len(v.Runs) != 0 {
		t.Fatal("validation failure created a run")
	}
	bad := opts
	bad.Window = 101
	if _, err := GrowTree(dir, nil, false, bad, io.Discard); err == nil {
		t.Fatal("invalid window accepted")
	}
	lock, err := treeLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := SelectTree(dir, []string{"b000001"}); err == nil {
		t.Fatal("concurrent mutation accepted")
	}
	lock()
	// A forged snapshot is rejected before starting a run.
	snapshot := filepath.Join(nodeRound(dir, "b000001"), "w001.snapshot.json")
	original, err := os.ReadFile(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(snapshot, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := GrowTree(dir, nil, false, opts, io.Discard); err == nil {
		t.Fatal("changed source accepted")
	}
	if err := os.WriteFile(snapshot, original, 0644); err != nil {
		t.Fatal(err)
	}
	v, _ = ReadTree(dir)
	if len(v.Runs) != 0 {
		t.Fatal("invalid source consumed compute")
	}
	response.RequestID = v.Nodes[0].RequestID
	if err := writeJSON(filepath.Join(nodeRound(dir, "b000001"), "response.json"), response); err != nil {
		t.Fatal(err)
	}
	// Progress I/O failure must leave a failed run, not publish successful children.
	if _, err := GrowTree(dir, nil, true, opts, failingWriter{}); err == nil {
		t.Fatal("ignored failed output")
	}
	v, err = ReadTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Nodes) != 1 || len(v.Runs) != 1 || v.Runs[0].Status != "failed" || v.Selected[0] != "b000001" {
		t.Fatal("failure advertised as successful")
	}
	// Recovery uses a new run ID and immutable parent.
	children, err := GrowTree(dir, nil, false, opts, io.Discard)
	if err != nil || len(children) != 1 {
		t.Fatal("cannot retry", err)
	}
}

type failingWriter struct{}

func TestTreePartialGenerationRecoveryAndTampering(t *testing.T) {
	input, _, _, response := fixture(t)
	dir := filepath.Join(t.TempDir(), "tree")
	if err := InitTree(input, dir, "", 200); err != nil {
		t.Fatal(err)
	}
	v, _ := ReadTree(dir)
	response.RequestID = v.Nodes[0].RequestID
	if err := writeJSON(filepath.Join(nodeRound(dir, "b000001"), "response.json"), response); err != nil {
		t.Fatal(err)
	}
	opts := TrialOptions{Ticks: 100, Every: 20, Window: 80, Workers: 2}
	parents, err := GrowTree(dir, nil, true, opts, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	children, err := GrowTree(dir, nil, false, opts, &limitedWriter{remaining: 2})
	if err == nil || len(children) != 1 {
		t.Fatal("expected one committed cohort and a later failure", children, err)
	}
	v, err = ReadTree(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Nodes) != 4 || !reflect.DeepEqual(v.Selected, parents) || v.Runs[1].Status != "failed" || len(v.Runs[1].Children) != 1 {
		t.Fatal("partial run lost provenance or changed selection")
	}
	// Cached numbers and ancestry must be backed by archived evidence and trial.
	path := filepath.Join(dir, "nodes", children[0], "node.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var n TreeNode
	if err := readJSON(path, &n); err != nil {
		t.Fatal(err)
	}
	n.Worlds[0].Diversity++
	if err := replaceJSON(path, n); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTree(dir); err == nil {
		t.Fatal("forged comparison metric accepted")
	}
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(path, &n); err != nil {
		t.Fatal(err)
	}
	n.Parent = parents[1]
	if err := replaceJSON(path, n); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadTree(dir); err == nil {
		t.Fatal("false ancestry accepted")
	}
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}
	// Export must not overwrite physical artifacts through a mistaken output path.
	if err := ExportTree(dir, filepath.Join(dir, "tree.json")); err == nil {
		t.Fatal("report replaced metadata")
	}
	children, err = GrowTree(dir, nil, false, opts, io.Discard)
	if err != nil || len(children) != 2 {
		t.Fatal("partial generation cannot be retried", err)
	}
}

type limitedWriter struct{ remaining int }

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.remaining == 0 {
		return 0, io.ErrClosedPipe
	}
	w.remaining--
	return len(p), nil
}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestTreeDoesNotDependOnOriginalBatch(t *testing.T) {
	input, _, _, _ := fixture(t)
	dir := filepath.Join(t.TempDir(), "tree")
	if err := InitTree(input, dir, "", 200); err != nil {
		t.Fatal(err)
	}
	// Rename within the test sandbox instead of deleting any input.
	if err := os.Rename(input, input+"-moved"); err != nil {
		t.Fatal(err)
	}
	if _, err := GrowTree(dir, nil, false, TrialOptions{Ticks: 100, Every: 20, Window: 80, Workers: 2}, io.Discard); err != nil {
		t.Fatal(err)
	}
}
