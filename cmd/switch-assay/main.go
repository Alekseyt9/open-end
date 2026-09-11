// switch-assay compares metabolic switching costs in evolving worlds.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
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
		fmt.Fprintln(os.Stderr, "switch-assay:", err)
		os.Exit(1)
	}
}
func read(path string, dest any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
func save(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
func run(args []string, out io.Writer) (err error) {
	fs := flag.NewFlagSet("switch-assay", flag.ContinueOnError)
	fs.SetOutput(out)
	input := fs.String("input", "", "completed experiment directory")
	dest := fs.String("out", "", "new output directory")
	condition := fs.String("case", "environment", "source case to compare")
	ticks := fs.Int("ticks", 20000, "additional ticks in each arm")
	every := fs.Int("every", 1000, "report interval")
	groupAge := fs.Uint64("group-age", 100, "minimum unchanged membership age")
	workers := fs.Int("workers", 16, "parallel worlds")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if *input == "" || *dest == "" || *ticks < 1 || *every < 1 || *every > *ticks || *workers < 1 || *workers > 256 || *groupAge == 0 || fs.NArg() != 0 {
		return fmt.Errorf("provide input/out, ticks >= every > 0, and workers 1..256")
	}
	var manifest struct {
		Status           string
		Completed, Total int
	}
	if err := read(filepath.Join(*input, "manifest.json"), &manifest); err != nil {
		return err
	}
	if manifest.Status != "complete" || manifest.Total < 1 || manifest.Total != manifest.Completed {
		return fmt.Errorf("source batch is incomplete")
	}
	var rows []struct {
		Case                  string
		Seed, Tick            uint64
		Snapshot, StateSHA256 string
	}
	if err := read(filepath.Join(*input, "summary.json"), &rows); err != nil {
		return err
	}
	if len(rows) != manifest.Total {
		return fmt.Errorf("source manifest/summary mismatch")
	}
	type job struct {
		index  int
		source *world.World
		mode   string
	}
	jobs := []job{}
	seen := map[uint64]bool{}
	for _, row := range rows {
		if row.Case != *condition {
			continue
		}
		if row.Snapshot == "" || filepath.Base(row.Snapshot) != row.Snapshot || seen[row.Seed] {
			return fmt.Errorf("invalid source snapshot or duplicate seed")
		}
		seen[row.Seed] = true
		f, e := os.Open(filepath.Join(*input, row.Snapshot))
		if e != nil {
			return e
		}
		w, e := kernel.Load(f)
		f.Close()
		if e != nil {
			return e
		}
		if w.Config.Environment == "" || w.Config.CollectiveAblation != "" || w.Config.BondMotion != "" || w.Config.MetabolicSwitchCost != 0 || w.RuleState != nil || kernel.Hash(w) != row.StateSHA256 || w.Tick != row.Tick || w.Config.Seed != row.Seed {
			return fmt.Errorf("source snapshot metadata mismatch")
		}
		for _, mode := range []string{"0", "1", "2", "4"} {
			jobs = append(jobs, job{len(jobs), w, mode})
		}
	}
	if len(jobs) == 0 {
		return fmt.Errorf("no source worlds match case %q", *condition)
	}
	if err := os.Mkdir(*dest, 0755); err != nil {
		return err
	}
	start := time.Now()
	status := map[string]any{"status": "running", "input": *input, "case": *condition, "workers": *workers, "ticks": *ticks, "every": *every, "group_age": *groupAge, "total": len(jobs), "started_utc": start.UTC().Format(time.RFC3339)}
	if err := save(filepath.Join(*dest, "manifest.json"), status); err != nil {
		return err
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
	type result struct {
		trial experiment.SwitchTrial
		err   error
	}
	queue := make(chan job, len(jobs))
	results := make(chan result, len(jobs))
	for _, j := range jobs {
		queue <- j
	}
	close(queue)
	var wg sync.WaitGroup
	for i := 0; i < min(*workers, len(jobs)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				w, frames, r, e := experiment.ContinueSwitch(j.source, j.mode, *ticks, *every, *groupAge)
				if e == nil {
					stem := fmt.Sprintf("w%03d-%s", j.index/4+1, j.mode)
					r.Case = *condition + "-" + j.mode
					r.Snapshot = stem + ".snapshot.json"
					r.Metrics = stem + ".jsonl"
					f, openErr := os.Create(filepath.Join(*dest, r.Snapshot))
					e = openErr
					if e == nil {
						e = kernel.Save(f, w)
						closeErr := f.Close()
						if e == nil {
							e = closeErr
						}
					}
					if e == nil {
						f, e = os.Create(filepath.Join(*dest, r.Metrics))
						if e == nil {
							enc := json.NewEncoder(f)
							for _, m := range frames {
								if e = enc.Encode(m); e != nil {
									break
								}
							}
							closeErr := f.Close()
							if e == nil {
								e = closeErr
							}
						}
					}
				}
				results <- result{r, e}
			}
		}()
	}
	wg.Wait()
	close(results)
	trials := []experiment.SwitchTrial{}
	for result := range results {
		if result.err != nil {
			if err == nil {
				err = result.err
			}
		} else {
			trials = append(trials, result.trial)
		}
	}
	if err != nil {
		return err
	}
	sort.Slice(trials, func(i, j int) bool {
		if trials[i].Seed != trials[j].Seed {
			return trials[i].Seed < trials[j].Seed
		}
		return trials[i].Variant < trials[j].Variant
	})
	if err = save(filepath.Join(*dest, "results.json"), trials); err != nil {
		return err
	}
	status["status"] = "complete"
	summary := make([]map[string]any, 0, len(trials))
	for _, r := range trials {
		summary = append(summary, map[string]any{"Case": r.Case, "Seed": r.Seed, "Tick": r.Summary.ToTick, "Entities": r.Summary.Population.End, "Genomes": r.Summary.Genomes.End, "Snapshot": r.Snapshot, "Metrics": r.Metrics, "StateSHA256": r.FinalHash})
	}
	if err = save(filepath.Join(*dest, "summary.json"), summary); err != nil {
		return err
	}
	status["completed"] = len(trials)
	status["seconds"] = time.Since(start).Seconds()
	if err = save(filepath.Join(*dest, "manifest.json"), status); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Compared %d matched snapshots in %.2f seconds: %s\n", len(trials)/4, status["seconds"], filepath.Join(*dest, "results.json"))
	return err
}
