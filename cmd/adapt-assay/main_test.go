package main

import (
	"io"
	"open-end/internal/kernel"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParallelAssaySelectionAndNormalContinuation(t *testing.T) {
	input := t.TempDir()
	rows := []row{}
	sources := map[uint64]*world.World{}
	for seed := uint64(1); seed <= 3; seed++ {
		c := world.DefaultConfig()
		c.Width = 8
		c.Height = 8
		c.MaxEntities = 64
		c.MutationPPM = 0
		c.Seed = seed
		w, e := world.New(c)
		if e != nil {
			t.Fatal(e)
		}
		sources[seed] = w
		name := string(rune('a'+seed)) + ".json"
		f, e := os.Create(filepath.Join(input, name))
		if e != nil {
			t.Fatal(e)
		}
		e = kernel.Save(f, w)
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
		rows = append(rows, row{Case: "symbols", Seed: seed, Snapshot: name, StateSHA256: kernel.Hash(w)})
	}
	if e := save(filepath.Join(input, "summary.json"), rows); e != nil {
		t.Fatal(e)
	}
	if e := save(filepath.Join(input, "manifest.json"), map[string]any{"Status": "complete", "Total": 3, "Completed": 3}); e != nil {
		t.Fatal(e)
	}
	var reference []evaluation
	for _, workers := range []string{"1", "16"} {
		dir := filepath.Join(t.TempDir(), "assay")
		args := []string{"-input", input, "-out", dir, "-workers", workers, "-ticks", "30", "-continue-ticks", "20", "-every", "10", "-select", "2"}
		if e := run(args, io.Discard); e != nil {
			t.Fatal(e)
		}
		var es []evaluation
		if e := read(filepath.Join(dir, "evaluation.json"), &es); e != nil {
			t.Fatal(e)
		}
		if workers == "1" {
			reference = es
		} else if !reflect.DeepEqual(es, reference) {
			t.Fatal("worker count changes evaluation")
		}
		var final []row
		if e := read(filepath.Join(dir, "summary.json"), &final); e != nil {
			t.Fatal(e)
		}
		if len(final) != 2 {
			t.Fatal("wrong continuation count")
		}
		for _, r := range final {
			w := sources[r.Seed]
			for i := 0; i < 20; i++ {
				kernel.Step(w)
			}
			if kernel.Hash(w) != r.StateSHA256 {
				t.Fatal("probe contaminated normal continuation")
			}
			f, e := os.Open(filepath.Join(input, rows[r.Seed-1].Snapshot))
			if e != nil {
				t.Fatal(e)
			}
			w, e = kernel.Load(f)
			f.Close()
			if e != nil {
				t.Fatal(e)
			}
			sources[r.Seed] = w
		}
		if e := run(args, io.Discard); e == nil {
			t.Fatal("existing output overwritten")
		}
	}
	rows[0].StateSHA256 = "tampered"
	if e := save(filepath.Join(input, "summary.json"), rows); e != nil {
		t.Fatal(e)
	}
	if e := run([]string{"-input", input, "-out", filepath.Join(t.TempDir(), "bad"), "-select", "2"}, io.Discard); e == nil {
		t.Fatal("bad source accepted")
	}
}
