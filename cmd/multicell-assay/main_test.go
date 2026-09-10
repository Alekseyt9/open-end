package main

import (
	"io"
	"open-end/internal/council"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"testing"
)

func TestMatchedAssayWorkersAndSourceIntegrity(t *testing.T) {
	input := t.TempDir()
	rows := []map[string]any{}
	for seed := uint64(1); seed <= 2; seed++ {
		c := world.DefaultConfig()
		c.Width = 8
		c.Height = 8
		c.MaxEntities = 64
		c.Ecology = true
		c.Environment = "coupled"
		c.CopyModel = "evolving"
		c.Seed = seed
		w, err := world.New(c)
		if err != nil {
			t.Fatal(err)
		}
		name := string(rune('a'+seed)) + ".json"
		f, err := os.Create(filepath.Join(input, name))
		if err != nil {
			t.Fatal(err)
		}
		err = kernel.Save(f, w)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, map[string]any{"Case": "environment", "Seed": seed, "Tick": 0, "Snapshot": name, "StateSHA256": kernel.Hash(w)})
	}
	if err := save(filepath.Join(input, "manifest.json"), map[string]any{"Status": "complete", "Total": 2, "Completed": 2}); err != nil {
		t.Fatal(err)
	}
	if err := save(filepath.Join(input, "summary.json"), rows); err != nil {
		t.Fatal(err)
	}
	var hashes []string
	for _, workers := range []string{"1", "16"} {
		dir := filepath.Join(t.TempDir(), "assay")
		args := []string{"-input", input, "-out", dir, "-ticks", "100", "-every", "20", "-workers", workers}
		if err := run(args, io.Discard); err != nil {
			t.Fatal(err)
		}
		var results []experiment.MulticellTrial
		if err := read(filepath.Join(dir, "results.json"), &results); err != nil {
			t.Fatal(err)
		}
		if len(results) != 4 {
			t.Fatal("missing arms")
		}
		for i, r := range results {
			if workers == "1" {
				hashes = append(hashes, r.FinalHash)
			} else if hashes[i] != r.FinalHash {
				t.Fatal("workers change physics")
			}
		}
		if err := run(args, io.Discard); err == nil {
			t.Fatal("existing output overwritten")
		}
		if workers == "16" {
			tree := filepath.Join(t.TempDir(), "tree")
			if err := council.InitTree(dir, tree, "", 100); err != nil {
				t.Fatal(err)
			}
			if _, err := council.GrowTree(tree, nil, false, council.TrialOptions{Ticks: 100, Every: 20, Window: 80, Workers: 16}, io.Discard); err != nil {
				t.Fatal(err)
			}
			v, err := council.ReadTree(tree)
			if err != nil {
				t.Fatal(err)
			}
			for _, n := range v.Nodes {
				if len(n.Worlds) != 4 {
					t.Fatal("tree lost arms")
				}
				for _, w := range n.Worlds {
					if w.CollectiveAge != 100 {
						t.Fatal("tree lost observation protocol")
					}
				}
			}
		}
	}
	rows[0]["StateSHA256"] = "tampered"
	if err := save(filepath.Join(input, "summary.json"), rows); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-input", input, "-out", filepath.Join(t.TempDir(), "bad")}, io.Discard); err == nil {
		t.Fatal("source hash mismatch accepted")
	}
}
