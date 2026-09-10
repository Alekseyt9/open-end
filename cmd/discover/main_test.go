package main

import (
	"io"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReplayCLIProvenanceWorkersAndFailureManifest(t *testing.T) {
	source, input := t.TempDir(), t.TempDir()
	c := world.DefaultConfig()
	c.Width, c.Height, c.MaxEntities = 8, 8, 64
	c.Ecology, c.Environment, c.CopyModel = true, "coupled", "evolving"
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	writeWorld := func(dir, name string, w *world.World) {
		t.Helper()
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		err = kernel.Save(f, w)
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
	}
	writeJSON := func(dir, name string, v any) {
		t.Helper()
		if _, err := save(filepath.Join(dir, name), v); err != nil {
			t.Fatal(err)
		}
	}
	writeWorld(source, "source.json", w)
	writeJSON(source, "summary.json", []map[string]any{{"Case": "environment", "Seed": w.Config.Seed, "Tick": w.Tick, "Snapshot": "source.json", "StateSHA256": kernel.Hash(w)}})
	writeJSON(source, "manifest.json", manifest{Status: "complete", Total: 1, Completed: 1})
	trials := []experiment.EnvironmentTrial{}
	for _, mode := range []string{"intact", "bonds", "sharing", "signal-reading"} {
		end, _, r, err := experiment.ContinueCollectives(w, mode, 100, 20, 10)
		if err != nil {
			t.Fatal(err)
		}
		r.Snapshot = mode + ".json"
		writeWorld(input, r.Snapshot, end)
		trials = append(trials, r)
	}
	writeJSON(input, "results.json", trials)
	writeJSON(input, "manifest.json", manifest{Status: "complete", Total: 4, Completed: 4, Ticks: 100})
	var baseline []result
	for _, workers := range []string{"1", "16"} {
		dest := filepath.Join(t.TempDir(), "discovery")
		args := []string{"-input", input, "-source", source, "-out", dest, "-workers", workers, "-every", "20", "-group-age", "10"}
		if err := run(args, io.Discard); err != nil {
			t.Fatal(err)
		}
		var index []result
		if err := read(filepath.Join(dest, "index.json"), &index); err != nil {
			t.Fatal(err)
		}
		if len(index) != 4 {
			t.Fatal("lost treatments")
		}
		if workers == "1" {
			baseline = index
		} else if !reflect.DeepEqual(baseline, index) {
			t.Fatal("worker count changes analysis or report hashes")
		}
		if err := run(args, io.Discard); err == nil {
			t.Fatal("existing output overwritten")
		}
	}
	trials[0].InitialHash = "tampered"
	writeJSON(input, "results.json", trials)
	dest := filepath.Join(t.TempDir(), "bad")
	args := []string{"-input", input, "-source", source, "-out", dest}
	if err := run(args, io.Discard); err == nil {
		t.Fatal("bad initial hash accepted")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("output created before provenance validation")
	}
	trials[0].InitialHash = trials[0].SourceHash
	// A valid but unrelated final snapshot passes file validation, then fails replay.
	other, _, _, err := experiment.ContinueCollectives(w, "intact", 101, 20, 10)
	if err != nil {
		t.Fatal(err)
	}
	other.Tick = 100 // Valid state at the claimed tick, but different from the replay.
	writeWorld(input, trials[0].Snapshot, other)
	trials[0].FinalHash = kernel.Hash(other)
	writeJSON(input, "results.json", trials)
	if err := run(args, io.Discard); err == nil {
		t.Fatal("divergent replay published as success")
	}
	var m manifest
	if err := read(filepath.Join(dest, "manifest.json"), &m); err != nil {
		t.Fatal(err)
	}
	if m.Status != "failed" {
		t.Fatal("failed replay marked complete")
	}
}
