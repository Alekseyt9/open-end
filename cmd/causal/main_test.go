package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"open-end/internal/discovery"
	"open-end/internal/evolution"
	"open-end/internal/kernel"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCLIWorkerReproducibilityAndEvidenceValidation(t *testing.T) {
	input, snapshots := t.TempDir(), t.TempDir()
	index := []map[string]any{}
	rows := []map[string]any{}
	write := func(dir, name string, value any) string {
		t.Helper()
		h, e := save(filepath.Join(dir, name), value)
		if e != nil {
			t.Fatal(e)
		}
		return h
	}
	for seed := uint64(1); seed <= 3; seed++ {
		c := world.DefaultConfig()
		c.Width, c.Height, c.MaxEntities = 8, 8, 64
		c.MutationPPM = 0
		c.Seed = seed
		w, e := world.New(c)
		if e != nil {
			t.Fatal(e)
		}
		p := w.Particles[1]
		p.Code = []vm.Instruction{{Op: vm.NOP}}
		p.Origin = evolution.Hash(p.Code, p.InitialMemory)
		w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
		w.RegisterGenome(p, "")
		for _, i := range []vm.Instruction{{Op: vm.ALLOCATE, A: 1}, {Op: vm.COPY}, {Op: vm.BIND}} {
			rules.Resolve(w, rules.Event{Actor: 1, Intent: vm.Intent{Instruction: i}})
		}
		w.Tick = 100
		name := string(rune('a'+seed)) + ".json"
		f, e := os.Create(filepath.Join(snapshots, name))
		if e != nil {
			t.Fatal(e)
		}
		e = kernel.Save(f, w)
		closeErr := f.Close()
		if e != nil || closeErr != nil {
			t.Fatal(e, closeErr)
		}
		ids := []uint64{1, 2}
		b, _ := json.Marshal(ids)
		sum := sha256.Sum256(b)
		id := hex.EncodeToString(sum[:])
		r := discovery.Report{Version: 1, Seed: seed, Treatment: "intact", Config: discovery.Config{MinAge: 10}, FinalHash: kernel.Hash(w), Frames: []discovery.Frame{{Tick: 100, MicroIDs: ids, Nodes: []discovery.Node{{ID: id, Members: ids, Sources: []string{"bonds"}, PersistentBond: true, BondAge: 100}}}}}
		h := write(input, name, r)
		index = append(index, map[string]any{"seed": seed, "treatment": "intact", "report": name, "report_sha256": h, "verified_final_sha256": r.FinalHash})
		rows = append(rows, map[string]any{"Seed": seed, "Tick": 100, "Snapshot": name, "StateSHA256": r.FinalHash})
	}
	for _, dir := range []string{input, snapshots} {
		write(dir, "manifest.json", map[string]any{"status": "complete", "total": 3, "completed": 3})
	}
	write(input, "index.json", index)
	write(snapshots, "summary.json", rows)
	var baseline any
	for _, workers := range []string{"1", "16"} {
		dest := filepath.Join(t.TempDir(), "run")
		args := []string{"-input", input, "-snapshots", snapshots, "-out", dest, "-workers", workers, "-horizon", "2", "-blocks", "2", "-permutations", "1"}
		if e := run(args, io.Discard); e != nil {
			t.Fatal(e)
		}
		var v any
		if e := read(filepath.Join(dest, "index.json"), &v); e != nil {
			t.Fatal(e)
		}
		if workers == "1" {
			baseline = v
		} else if !reflect.DeepEqual(baseline, v) {
			t.Fatal("workers changed outputs or prediction hashes")
		}
		if e := run(args, io.Discard); e == nil {
			t.Fatal("existing directory overwritten")
		}
	}
	index[0]["report_sha256"] = "tampered"
	write(input, "index.json", index)
	dest := filepath.Join(t.TempDir(), "bad")
	if e := run([]string{"-input", input, "-snapshots", snapshots, "-out", dest}, io.Discard); e == nil {
		t.Fatal("corrupt discovery accepted")
	}
	if _, e := os.Stat(dest); !os.IsNotExist(e) {
		t.Fatal("output created before input validation")
	}
}
