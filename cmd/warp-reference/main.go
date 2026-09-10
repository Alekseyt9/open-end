// warp-reference exports validated fixtures and times the production Go kernel.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"open-end/internal/dsl"
	"open-end/internal/evolution"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
	"open-end/internal/vm"
	"open-end/internal/world"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"
)

type physicalParticle struct {
	ID            uint64 `json:"id"`
	Parent        uint64 `json:"parent"`
	Position      int    `json:"position"`
	Energy        int    `json:"energy"`
	Code          any    `json:"code"`
	Memory        [8]int `json:"memory"`
	InitialMemory [8]int `json:"initial_memory"`
	IP            int    `json:"ip"`
	Flag          bool   `json:"flag"`
	Target        uint64 `json:"target"`
	Created       uint64 `json:"created"`
	Generation    uint64 `json:"generation"`
}

func project(w *world.World) map[string]any {
	particles := []physicalParticle{}
	for _, p := range w.Particles {
		particles = append(particles, physicalParticle{p.ID, p.Parent, p.Position, p.Energy, p.Code, p.Memory, p.InitialMemory, p.IP, p.Flag, p.Target, p.Created, p.Generation})
	}
	sort.Slice(particles, func(i, j int) bool { return particles[i].ID < particles[j].ID })
	relations := []world.Relation{}
	for _, r := range w.Relations {
		relations = append(relations, r)
	}
	sort.Slice(relations, func(i, j int) bool {
		if relations[i].A != relations[j].A {
			return relations[i].A < relations[j].A
		}
		return relations[i].B < relations[j].B
	})
	out := map[string]any{"tick": w.Tick, "rng": w.RNG.State, "transport_rng": w.TransportRNG.State, "next_id": w.NextID, "cells": w.Cells, "particles": particles, "relations": relations, "accounting": w.Accounting}
	if w.RuleState != nil {
		out["rule_state"] = w.RuleState
	}
	return out
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	size := flag.Int("size", 32, "square grid size")
	count := flag.Int("worlds", 16, "independent worlds with seeds 1..N")
	ticks := flag.Int("ticks", 1000, "measured ticks")
	warm := flag.Int("warmup", 0, "unmeasured initial ticks")
	workers := flag.Int("workers", 16, "CPU workers / GOMAXPROCS")
	scenario := flag.String("scenario", "ecology", "baseline, ecology, mutation, dsl, assay, odd, opcodes")
	input := flag.String("input", "", "reuse initial worlds from a reference fixture")
	snapshotPath := flag.String("snapshot", "", "load a validated Go snapshot as a single world")
	output := flag.String("output", "", "output fixture JSON")
	flag.Parse()
	if *ticks < 0 || *warm < 0 || *count < 1 || *count > 65536 || *workers < 1 || *workers > 256 || *output == "" {
		return fmt.Errorf("invalid benchmark arguments")
	}
	runtime.GOMAXPROCS(*workers)
	worlds := []*world.World{}
	if *snapshotPath != "" {
		if *input != "" {
			return fmt.Errorf("snapshot and input are mutually exclusive")
		}
		f, err := os.Open(*snapshotPath)
		if err != nil {
			return err
		}
		w, err := kernel.Load(f)
		f.Close()
		if err != nil {
			return err
		}
		worlds = append(worlds, w)
	} else if *input != "" {
		b, err := os.ReadFile(*input)
		if err != nil {
			return err
		}
		var fixture struct {
			Initial []*world.World `json:"initial"`
		}
		if err := json.Unmarshal(b, &fixture); err != nil {
			return err
		}
		if len(fixture.Initial) == 0 {
			return fmt.Errorf("fixture contains no worlds")
		}
		worlds = fixture.Initial
	} else {
		for i := 0; i < *count; i++ {
			c := world.DefaultConfig()
			c.Width = *size
			c.Height = *size
			c.MaxEntities = c.Width * c.Height
			c.Seed = uint64(i + 1)
			c.MatterDiffusion = 4
			c.ChemicalDiffusion = 4
			c.Ecology = true
			c.MutationPPM = 10000
			switch *scenario {
			case "baseline":
				c.Ecology = false
			case "mutation":
				c.MutationPPM = 100000
			case "odd":
				c.Height = *size + 1
				c.MaxEntities = c.Width * c.Height
				c.MatterDiffusion = 3
				c.ChemicalDiffusion = 5
				c.MutationPPM = 100000
			case "ecology", "dsl", "assay", "opcodes":
			default:
				return fmt.Errorf("unknown scenario")
			}
			w, err := world.New(c)
			if err != nil {
				return err
			}
			if *scenario == "opcodes" {
				opcodeFixture(w, i)
			}
			if *scenario == "assay" {
				if *size != 32 {
					return fmt.Errorf("assay requires size 32")
				}
				b, err := os.ReadFile("experiments/ecology/isolation/pair.json")
				if err != nil {
					return err
				}
				var pair experiment.Pair
				if err := json.Unmarshal(b, &pair); err != nil {
					return err
				}
				w, err = experiment.Inoculate(pair, uint64(i+1), "mixed", 32)
				if err != nil {
					return err
				}
			}
			if *scenario == "dsl" {
				for j, name := range []string{"baseline", "direct-x"} {
					f, err := os.Open("examples/rules/" + name + ".json")
					if err != nil {
						return err
					}
					m, err := dsl.Parse(f)
					f.Close()
					if err != nil {
						return err
					}
					if j == 0 {
						err = kernel.ReloadRules(w, m)
					} else {
						err = kernel.ScheduleRules(w, dsl.Change{Tick: 100, Module: m})
					}
					if err != nil {
						return err
					}
				}
				if err := kernel.ScheduleRules(w, dsl.Change{Tick: 200, Rollback: true}); err != nil {
					return err
				}
			}
			worlds = append(worlds, w)
		}
	}
	for _, w := range worlds {
		if w != nil && w.Config.Symbols != "" {
			return fmt.Errorf("Warp does not support symbol inscriptions; use the Go simulator")
		}
		if w != nil && w.Config.CollectiveAblation != "" {
			return fmt.Errorf("Warp does not support collective ablations; use Go")
		}
		if w != nil && w.Config.Environment != "" {
			return fmt.Errorf("Warp does not support environment engineering; use the Go simulator")
		}
		if w != nil && w.Config.CopyModel != "" {
			return fmt.Errorf("Warp does not support copy-model %q; use the Go simulator", w.Config.CopyModel)
		}
		if err := w.Validate(); err != nil {
			return err
		}
		for i := 0; i < *warm; i++ {
			kernel.Step(w)
		}
	}
	initial, err := json.Marshal(worlds)
	if err != nil {
		return err
	}
	jobs := make(chan *world.World, len(worlds))
	for _, w := range worlds {
		jobs <- w
	}
	close(jobs)
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < min(*workers, len(worlds)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range jobs {
				for tick := 0; tick < *ticks; tick++ {
					kernel.Step(w)
				}
			}
		}()
	}
	wg.Wait()
	seconds := time.Since(start).Seconds()
	final := []map[string]any{}
	for _, w := range worlds {
		if err := w.Validate(); err != nil {
			return err
		}
		final = append(final, project(w))
	}
	f, err := os.Create(*output)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(map[string]any{"initial": json.RawMessage(initial), "final": final, "seconds": seconds, "ticks": *ticks, "workers": *workers, "go": runtime.Version(), "kernel": kernel.Version, "rules": kernel.RuleVersion})
}

// Isolated fixtures ensure all VM operations, including uncommon relation and
// memory instructions, are exercised independently of evolutionary luck.
func opcodeFixture(w *world.World, variant int) {
	p := w.Particles[1]
	p.Memory[0] = 150
	p.Memory[7] = -9
	op := vm.Opcode(variant % int(vm.EcologyOpcodeCount))
	a, b := -1, 0
	switch op {
	case vm.SENSE:
		a, b = 3, 7
	case vm.ABSORB, vm.TRANSFER, vm.TAKE:
		a = 48
	case vm.WRITE:
		a, b = -1, 123
	case vm.READ:
		a, b = -1, 2
	case vm.COMPARE:
		a, b = 0, 100
	case vm.JUMP:
		a, b = -1, -1
	case vm.CONVERT:
		a, b = (variant/int(vm.EcologyOpcodeCount))%2, 8
	}
	p.Code = []vm.Instruction{{Op: op, A: a, B: b}, {Op: vm.NOP}, {Op: vm.JUMP, A: 0}}
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.Origins = map[string]*world.Origin{p.Origin: {Hash: p.Origin, Copies: 1}}
	w.Genomes = make(map[string]*world.GenomeRecord)
	w.RegisterGenome(p, "")
	pos := w.Neighbor(p.Position, 1)
	q := &world.Particle{ID: 2, Position: pos, Energy: 64}
	w.Particles[2] = q
	w.Cells[pos].Occupant = 2
	w.Cells[pos].Matter--
	w.NextID = 3
	p.Target = 2
	if op == vm.UNBIND {
		w.Relations[world.RelationKey(1, 2)] = world.Relation{A: 1, B: 2}
	}
	if variant >= int(vm.EcologyOpcodeCount) {
		p.Energy = 2
	}
	e, m := w.Totals()
	w.Accounting.InitialEnergy = e
	w.Accounting.InitialMatter = m
}
