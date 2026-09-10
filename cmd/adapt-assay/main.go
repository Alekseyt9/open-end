// adapt-assay evaluates information benefits and allocates continuation slots.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/adaptivity"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "adapt-assay:", err)
		os.Exit(1)
	}
}
func read(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func save(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}

type row struct {
	Case                           string
	Seed, Tick                     uint64
	Snapshot, Metrics, StateSHA256 string
	Entities, Genomes              int
}
type evaluation struct {
	adaptivity.Candidate
	SourceSnapshot    string `json:"source_snapshot"`
	BehaviorFinalHash string `json:"behavior_final_sha256"`
}

func run(args []string, out io.Writer) (err error) {
	fs := flag.NewFlagSet("adapt-assay", flag.ContinueOnError)
	fs.SetOutput(out)
	input := fs.String("input", "", "completed batch directory")
	dest := fs.String("out", "", "new output directory")
	condition := fs.String("case", "symbols", "source case")
	workers := fs.Int("workers", 16, "parallel worlds")
	slots := fs.Int("select", 4, "worlds to continue")
	continuation := fs.Int("continue-ticks", 20000, "ordinary ticks per selected world")
	every := fs.Int("every", 1000, "continuation reporting interval")
	seed := fs.Uint64("selection-seed", 1, "deterministic exploratory selection")
	c := adaptivity.DefaultConfig()
	fs.IntVar(&c.Ticks, "ticks", c.Ticks, "ticks per probe and behavior reference")
	fs.Float64Var(&c.MinBenefit, "min-benefit", c.MinBenefit, "minimum mean information benefit before assigning a score")
	if e := fs.Parse(args); e != nil {
		if e == flag.ErrHelp {
			return nil
		}
		return e
	}
	if *input == "" || *dest == "" || *workers < 1 || *workers > 256 || *slots < 1 || *continuation < 1 || *every < 1 || *every > *continuation || fs.NArg() != 0 {
		return fmt.Errorf("provide input/out, positive selection and continuation, valid interval and workers 1..256")
	}
	if e := c.Validate(); e != nil {
		return e
	}
	var manifest struct {
		Status           string
		Total, Completed int
	}
	if e := read(filepath.Join(*input, "manifest.json"), &manifest); e != nil {
		return e
	}
	if manifest.Status != "complete" || manifest.Total < 1 || manifest.Total != manifest.Completed {
		return fmt.Errorf("incomplete source batch")
	}
	var rows []row
	if e := read(filepath.Join(*input, "summary.json"), &rows); e != nil {
		return e
	}
	if len(rows) != manifest.Total {
		return fmt.Errorf("source manifest mismatch")
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Seed < rows[j].Seed })
	sources := []*world.World{}
	evaluations := []evaluation{}
	seen := map[uint64]bool{}
	var context []byte
	var tick uint64
	for _, r := range rows {
		if r.Case != *condition {
			continue
		}
		if r.Snapshot == "" || filepath.Base(r.Snapshot) != r.Snapshot || seen[r.Seed] {
			return fmt.Errorf("invalid snapshot or duplicate seed")
		}
		seen[r.Seed] = true
		f, e := os.Open(filepath.Join(*input, r.Snapshot))
		if e != nil {
			return e
		}
		w, e := kernel.Load(f)
		f.Close()
		if e != nil {
			return e
		}
		if kernel.Hash(w) != r.StateSHA256 || w.Tick != r.Tick || w.Config.Seed != r.Seed || adaptivity.Executable(w) == 0 {
			return fmt.Errorf("source metadata mismatch or extinct source")
		}
		cfg := w.Config
		cfg.Seed = 0
		var active any
		if w.RuleState != nil {
			if len(w.RuleState.Pending) > 0 {
				return fmt.Errorf("scheduled transitions unsupported")
			}
			active = w.RuleState.Active
		}
		ctx, _ := json.Marshal(struct {
			Config world.Config
			Active any
		}{cfg, active})
		if len(sources) > 0 && (!bytes.Equal(ctx, context) || tick != w.Tick) {
			return fmt.Errorf("source contexts/horizons differ")
		}
		context = ctx
		tick = w.Tick
		sources = append(sources, w)
		evaluations = append(evaluations, evaluation{Candidate: adaptivity.Candidate{Seed: r.Seed, SourceHash: r.StateSHA256, Probes: make([]adaptivity.Probe, 12)}, SourceSnapshot: r.Snapshot})
	}
	if len(sources) < *slots {
		return fmt.Errorf("fewer matching worlds than selection slots")
	}
	if e := os.Mkdir(*dest, 0755); e != nil {
		return e
	}
	start := time.Now()
	status := map[string]any{"status": "running", "version": adaptivity.Version, "input": *input, "case": *condition, "config": c, "workers": *workers, "selected_slots": *slots, "selection_seed": *seed, "continue_ticks": *continuation, "every": *every, "probes": 12 * len(sources), "behavior_references": len(sources), "started_utc": start.UTC().Format(time.RFC3339), "go_version": runtime.Version()}
	if e := save(filepath.Join(*dest, "manifest.json"), status); e != nil {
		return e
	}
	defer func() {
		if err != nil {
			status["status"] = "failed"
			status["error"] = err.Error()
			_ = save(filepath.Join(*dest, "manifest.json"), status)
		}
	}()
	old := runtime.GOMAXPROCS(*workers)
	defer runtime.GOMAXPROCS(old)
	type job struct {
		World, Probe int
		Challenge    adaptivity.Challenge
		Mode         string
	}
	type result struct {
		job
		ProbeResult adaptivity.Probe
		Diversity   float64
		Descriptor  []float64
		Hash        string
		Err         error
	}
	queue := make(chan job, 13*len(sources))
	results := make(chan result, 13*len(sources))
	for i := range sources {
		j := 0
		for _, challenge := range adaptivity.Challenges() {
			for _, mode := range adaptivity.Modes {
				queue <- job{i, j, challenge, mode}
				j++
			}
		}
		queue <- job{World: i, Probe: -1}
	}
	close(queue)
	var wg sync.WaitGroup
	for worker := 0; worker < min(*workers, 13*len(sources)); worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				r := result{job: j}
				if j.Probe < 0 {
					w, e := adaptivity.Clone(sources[j.World])
					r.Err = e
					if e == nil {
						for k := 0; k < c.Ticks; k++ {
							kernel.Step(w)
						}
						r.Err = w.Validate()
						r.Diversity, r.Descriptor = adaptivity.FunctionalDiversity(sources[j.World], w)
						r.Hash = kernel.Hash(w)
					}
				} else {
					r.ProbeResult, r.Err = adaptivity.ProbeWorld(sources[j.World], c, j.Challenge, j.Mode)
				}
				results <- r
			}
		}()
	}
	wg.Wait()
	close(results)
	for r := range results {
		if r.Err != nil {
			if err == nil {
				err = r.Err
			}
			continue
		}
		e := &evaluations[r.World]
		if r.Probe < 0 {
			e.FunctionalDiversity = r.Diversity
			e.Descriptor = r.Descriptor
			e.BehaviorFinalHash = r.Hash
		} else {
			e.Probes[r.Probe] = r.ProbeResult
		}
	}
	if err != nil {
		return err
	}
	candidates := make([]adaptivity.Candidate, len(evaluations))
	for i := range evaluations {
		e := &evaluations[i]
		e.Selection, err = adaptivity.ScoreProbes(e.Probes, "selection", c)
		if err != nil {
			return err
		}
		e.Validation, err = adaptivity.ScoreProbes(e.Probes, "validation", c)
		if err != nil {
			return err
		}
		candidates[i] = e.Candidate
	}
	if err = adaptivity.Select(candidates, *slots, *seed); err != nil {
		return err
	}
	for i := range evaluations {
		evaluations[i].Candidate = candidates[i]
	}
	if err = save(filepath.Join(*dest, "evaluation.json"), evaluations); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(out, "Evaluated %d worlds with %d matched probes; continuing %d selected worlds.\n", len(sources), 12*len(sources), *slots); err != nil {
		return err
	}
	selected := make(chan int, *slots)
	type continued struct {
		Row row
		Err error
	}
	completed := make(chan continued, *slots)
	for i, c := range candidates {
		if c.Selected {
			selected <- i
		}
	}
	close(selected)
	for worker := 0; worker < min(*workers, *slots); worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range selected {
				stem := fmt.Sprintf("seed%d", sources[i].Config.Seed)
				r, e := continueWorld(sources[i], *dest, stem, *condition, *continuation, *every)
				completed <- continued{r, e}
			}
		}()
	}
	wg.Wait()
	close(completed)
	output := []row{}
	for r := range completed {
		if r.Err != nil {
			if err == nil {
				err = r.Err
			}
		} else {
			output = append(output, r.Row)
		}
	}
	if err != nil {
		return err
	}
	sort.Slice(output, func(i, j int) bool { return output[i].Seed < output[j].Seed })
	if err = save(filepath.Join(*dest, "summary.json"), output); err != nil {
		return err
	}
	status["status"] = "complete"
	status["total"] = len(output)
	status["completed"] = len(output)
	status["seconds"] = time.Since(start).Seconds()
	if err = save(filepath.Join(*dest, "manifest.json"), status); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Completed in %.2f seconds: %s\n", status["seconds"], filepath.Join(*dest, "evaluation.json"))
	return err
}

func continueWorld(source *world.World, dir, stem, condition string, ticks, every int) (r row, err error) {
	w, err := adaptivity.Clone(source)
	if err != nil {
		return r, err
	}
	r = row{Case: condition, Seed: w.Config.Seed, Snapshot: stem + ".snapshot.json", Metrics: stem + ".jsonl"}
	f, err := os.OpenFile(filepath.Join(dir, r.Metrics), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return r, err
	}
	defer func() {
		if e := f.Close(); err == nil {
			err = e
		}
	}()
	tracker := observer.NewTracker(w)
	enc := json.NewEncoder(f)
	if err = enc.Encode(tracker.Frame(w)); err != nil {
		return r, err
	}
	for i := 1; i <= ticks; i++ {
		kernel.StepObserved(w, tracker)
		if i%every == 0 || i == ticks {
			if err = enc.Encode(tracker.Frame(w)); err != nil {
				return r, err
			}
		}
	}
	state, err := os.OpenFile(filepath.Join(dir, r.Snapshot), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return r, err
	}
	err = kernel.Save(state, w)
	closeErr := state.Close()
	if err != nil {
		return r, err
	}
	if closeErr != nil {
		return r, closeErr
	}
	r.Tick = w.Tick
	r.StateSHA256 = kernel.Hash(w)
	r.Entities = len(w.Particles)
	r.Genomes = observer.Observe(w).Genomes
	return r, nil
}
