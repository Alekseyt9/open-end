package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/dsl"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "sim:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	c := world.DefaultConfig()
	fs := flag.NewFlagSet("sim", flag.ContinueOnError)
	fs.SetOutput(out)
	ticks := fs.Int("ticks", 10000, "number of additional ticks")
	every := fs.Int("every", 1000, "metrics interval in ticks")
	fs.Uint64Var(&c.Seed, "seed", c.Seed, "deterministic seed for new worlds")
	fs.IntVar(&c.Width, "width", c.Width, "grid width")
	fs.IntVar(&c.Height, "height", c.Height, "grid height")
	fs.IntVar(&c.MutationPPM, "mutation-ppm", c.MutationPPM, "mutation probability per COPY / COPYMEM, parts per million")
	fs.StringVar(&c.CopyModel, "copy-model", "", "copy policy model: evolving or fixed; empty preserves legacy physics")
	fs.IntVar(&c.Inflow, "inflow", c.Inflow, "maximum energy input per cell per tick")
	fs.IntVar(&c.MatterDiffusion, "matter-diffusion", c.MatterDiffusion, "conservative matter mixing interval; 0 disables")
	fs.BoolVar(&c.Ecology, "ecology", c.Ecology, "enable chemical reactions and ecology instruction mutations")
	fs.IntVar(&c.ChemicalDiffusion, "chemical-diffusion", c.ChemicalDiffusion, "chemical mixing interval; 0 disables")
	save := fs.String("save", "", "write final snapshot via a temporary file")
	load := fs.String("load", "", "resume snapshot including its config and RNG state")
	metricsPath := fs.String("metrics", "", "write JSONL metrics (new file only)")
	rulesPath := fs.String("rules", "", "initial JSON reaction module (requires ecology)")
	var changes ruleChanges
	fs.Var(&changes, "rule-change", "scheduled tick=JSON-path; repeatable, frozen in snapshot")
	rollback := fs.String("rollback-at", "", "tick at which to restore the previous rule module")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *ticks < 0 || *every < 1 {
		return fmt.Errorf("ticks must be nonnegative and every positive")
	}
	var w *world.World
	var err error
	if *load != "" {
		var conflict string
		fs.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "seed", "width", "height", "mutation-ppm", "inflow", "matter-diffusion", "ecology", "chemical-diffusion", "copy-model", "rules", "rule-change", "rollback-at":
				conflict = f.Name
			}
		})
		if conflict != "" {
			return fmt.Errorf("-%s cannot override snapshot configuration", conflict)
		}
		f, e := os.Open(*load)
		if e != nil {
			return e
		}
		w, err = kernel.Load(f)
		_ = f.Close()
	} else {
		c.MaxEntities = c.Width * c.Height
		w, err = world.New(c)
	}
	if err != nil {
		return err
	}
	if *rulesPath != "" {
		m, err := readRules(*rulesPath)
		if err != nil {
			return err
		}
		if err := kernel.ReloadRules(w, m); err != nil {
			return err
		}
	}
	var plan []dsl.Change
	for _, text := range changes {
		tickText, path, ok := strings.Cut(text, "=")
		tick, err := strconv.ParseUint(tickText, 10, 64)
		if !ok || err != nil || path == "" {
			return fmt.Errorf("invalid -rule-change: expected tick=path")
		}
		m, err := readRules(path)
		if err != nil {
			return err
		}
		plan = append(plan, dsl.Change{Tick: tick, Module: m})
	}
	if *rollback != "" {
		tick, err := strconv.ParseUint(*rollback, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid -rollback-at: %w", err)
		}
		plan = append(plan, dsl.Change{Tick: tick, Rollback: true})
	}
	sort.Slice(plan, func(i, j int) bool { return plan[i].Tick < plan[j].Tick })
	for _, change := range plan {
		if err := kernel.ScheduleRules(w, change); err != nil {
			return err
		}
	}
	var metrics io.Writer = io.Discard
	if *metricsPath != "" {
		if *save != "" {
			a, _ := filepath.Abs(*save)
			b, _ := filepath.Abs(*metricsPath)
			if a == b {
				return fmt.Errorf("snapshot and metrics paths must differ")
			}
		}
		if err := os.MkdirAll(filepath.Dir(*metricsPath), 0755); err != nil {
			return err
		}
		f, e := os.OpenFile(*metricsPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if e != nil {
			return e
		}
		defer f.Close()
		metrics = f
	}
	enc := json.NewEncoder(metrics)
	var tracker *observer.Tracker
	if *metricsPath != "" {
		tracker = observer.NewTracker(w)
	}
	previous := observer.Observe(w)
	report := func() error {
		var m observer.Metrics
		if tracker != nil {
			m = tracker.Frame(w)
			if !m.Telemetry.Complete {
				return fmt.Errorf("incomplete telemetry at tick %d", w.Tick)
			}
		} else {
			m = observer.Observe(w)
		}
		m.Interval = observer.Since(previous, m)
		previous = m
		if err := enc.Encode(m); err != nil {
			return err
		}
		_, err := fmt.Fprintf(out, "tick=%d entities=%d executable=%d genomes=%d lineages=%d copies=%d deaths=%d energy=%d new_copies=%d reactions=%v\n", m.Tick, m.Entities, m.Executable, m.Genomes, m.Lineages, m.Copies, m.Deaths, m.Energy, m.Interval.Copies, m.Converted)
		if err == nil && w.RuleState != nil {
			version, hash := "builtin", "builtin"
			if a := w.RuleState.Active; a != nil {
				version, hash = a.Source.Version, a.Hash
			}
			_, err = fmt.Fprintf(out, "rules=%s rules_sha256=%s rule_events=%d dsl_instructions=%d\n", version, hash, len(w.RuleState.Events), w.RuleState.Instructions)
		}
		return err
	}
	if err := report(); err != nil {
		return err
	}
	for t := 0; t < *ticks; t++ {
		if tracker != nil {
			kernel.StepObserved(w, tracker)
		} else {
			kernel.Step(w)
		}
		if (t+1)%*every == 0 || t+1 == *ticks {
			if err := report(); err != nil {
				return err
			}
		}
	}
	if err := w.Validate(); err != nil {
		return err
	}
	if *save != "" {
		if err := writeSnapshot(*save, w); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(out, "state_sha256=%s\n", kernel.Hash(w))
	return err
}

type ruleChanges []string

func (r *ruleChanges) String() string     { return strings.Join(*r, ",") }
func (r *ruleChanges) Set(s string) error { *r = append(*r, s); return nil }
func readRules(path string) (*dsl.Module, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return dsl.Parse(f)
}

func writeSnapshot(path string, w *world.World) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".snapshot-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err := kernel.Save(f, w); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
