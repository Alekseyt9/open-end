package main

import (
	"io"
	"open-end/internal/evolution"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRoleBatchWorkerEquivalenceAndIntegrity(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.MutationPPM = 0
	w, e := world.New(c)
	if e != nil {
		t.Fatal(e)
	}
	for _, i := range []vm.Instruction{{Op: vm.ALLOCATE, A: 1}, {Op: vm.COPY}, {Op: vm.BIND}} {
		rules.Resolve(w, rules.Event{Actor: 1, Intent: vm.Intent{Instruction: i}})
	}
	p := w.Particles[2]
	p.Code = []vm.Instruction{{Op: vm.ABSORB, A: 10}}
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
	w.RegisterGenome(p, "")
	input := t.TempDir()
	f, e := os.Create(filepath.Join(input, "source.json"))
	if e != nil {
		t.Fatal(e)
	}
	e = kernel.Save(f, w)
	f.Close()
	if e != nil {
		t.Fatal(e)
	}
	rows := []witness{{Case: "fixture", Seed: c.Seed, Snapshot: "source.json", Hash: kernel.Hash(w), Candidate: experiment.ClonalCandidate{Members: []uint64{1, 2}, Genomes: []observer.Lineage{{Hash: w.Particles[1].Genome, Count: 1}, {Hash: p.Genome, Count: 1}}}}}
	if e := save(filepath.Join(input, "witnesses.json"), rows); e != nil {
		t.Fatal(e)
	}
	if e := save(filepath.Join(input, "manifest.json"), map[string]any{"status": "complete", "total": 1, "completed": 1}); e != nil {
		t.Fatal(e)
	}
	var prior []result
	for _, workers := range []string{"1", "16"} {
		dir := filepath.Join(t.TempDir(), "output")
		args := []string{"-input", input, "-out", dir, "-ticks", "100", "-every", "20", "-workers", workers}
		if e := run(args, io.Discard); e != nil {
			t.Fatal(e)
		}
		var current []result
		if e := read(filepath.Join(dir, "results.json"), &current); e != nil {
			t.Fatal(e)
		}
		if len(current) != 7 {
			t.Fatal("missing arms")
		}
		if prior != nil && !reflect.DeepEqual(prior, current) {
			t.Fatal("worker count changes results")
		}
		prior = current
		if e := run(args, io.Discard); e == nil {
			t.Fatal("overwrote existing output")
		}
	}
	rows[0].Hash = "tampered"
	if e := save(filepath.Join(input, "witnesses.json"), rows); e != nil {
		t.Fatal(e)
	}
	if e := run([]string{"-input", input, "-out", filepath.Join(t.TempDir(), "bad")}, io.Discard); e == nil {
		t.Fatal("accepted altered source")
	}
}
