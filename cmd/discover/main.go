// discover replays a completed group assay and records candidate entity boundaries.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/discovery"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"sync"
	"time"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "discover:", err)
		os.Exit(1)
	}
}
func read(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
func save(path string, v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	b = append(b, '\n')
	if err = os.WriteFile(path, b, 0644); err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
func load(dir, name string) (*world.World, error) {
	if name == "" || filepath.Base(name) != name {
		return nil, fmt.Errorf("invalid snapshot name")
	}
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return kernel.Load(f)
}

type manifest struct {
	Status                  string
	Total, Completed, Ticks int
}

func completed(dir string) (manifest, error) {
	var m manifest
	err := read(filepath.Join(dir, "manifest.json"), &m)
	if err == nil && (m.Status != "complete" || m.Total < 1 || m.Completed != m.Total) {
		err = fmt.Errorf("incomplete batch %s", dir)
	}
	return m, err
}

type result struct {
	FlowReciprocal   int    `json:"flow_candidate_observations_with_next_reciprocal_connectivity"`
	Seed             uint64 `json:"seed"`
	Treatment        string `json:"treatment"`
	File             string `json:"report"`
	Hash             string `json:"report_sha256"`
	Source           string `json:"source_sha256"`
	Final            string `json:"verified_final_sha256"`
	Frames           int    `json:"frames"`
	EndCandidates    int    `json:"end_candidates"`
	PersistentBonds  int    `json:"end_persistent_bond_boundaries"`
	PersistentFlows  int    `json:"end_persistent_reciprocal_flows"`
	MaxDepth         int    `json:"maximum_structural_depth"`
	FlowObservations int    `json:"flow_candidate_observations_with_next_interval"`
	FlowRetained     int    `json:"flow_candidate_observations_with_next_internal_transfer"`
	Incomplete       int    `json:"incomplete_activity_intervals"`
	Daughters        int    `json:"daughter_candidates"`
}

func compact(r discovery.Report) result {
	v := result{Seed: r.Seed, Treatment: r.Treatment, Source: r.SourceHash, Final: r.FinalHash, Frames: len(r.Frames), Daughters: r.Daughters}
	for _, f := range r.Frames {
		if !f.Complete {
			v.Incomplete++
		}
		for _, n := range f.Nodes {
			v.MaxDepth = max(v.MaxDepth, n.Level)
			if slices.Contains(n.Sources, "reciprocal_transfer") && n.NextComplete && n.Next != nil {
				v.FlowObservations++
				if n.NextReciprocal {
					v.FlowReciprocal++
				}
				if n.Next.Internal > 0 {
					v.FlowRetained++
				}
			}
		}
	}
	last := r.Frames[len(r.Frames)-1]
	v.EndCandidates = len(last.Nodes)
	for _, n := range last.Nodes {
		if n.PersistentBond {
			v.PersistentBonds++
		}
		if n.PersistentFlow {
			v.PersistentFlows++
		}
	}
	return v
}
func run(args []string, out io.Writer) (err error) {
	c := discovery.DefaultConfig()
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	fs.SetOutput(out)
	input := fs.String("input", "", "completed group-assay directory")
	source := fs.String("source", "", "original completed source batch")
	dest := fs.String("out", "", "new output directory")
	workers := fs.Int("workers", 16, "independent replay workers")
	fs.IntVar(&c.Every, "every", c.Every, "discovery interval in ticks")
	fs.Uint64Var(&c.MinAge, "group-age", c.MinAge, "minimum boundary age")
	fs.IntVar(&c.Limit, "record-limit", c.Limit, "per-interval directed-edge and actor limits")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if *input == "" || *source == "" || *dest == "" || *workers < 1 || *workers > 256 || c.Every < 1 || c.MinAge == 0 || c.Limit < 1 || c.Limit > 1000000 || fs.NArg() != 0 {
		return fmt.Errorf("provide input/source/out, positive observation settings and workers 1..256")
	}
	im, err := completed(*input)
	if err != nil {
		return err
	}
	sm, err := completed(*source)
	if err != nil {
		return err
	}
	var rows []struct {
		Case, Snapshot, StateSHA256 string
		Seed, Tick                  uint64
	}
	if err := read(filepath.Join(*source, "summary.json"), &rows); err != nil {
		return err
	}
	if len(rows) != sm.Total {
		return fmt.Errorf("source manifest/summary mismatch")
	}
	var trials []experiment.EnvironmentTrial
	if err := read(filepath.Join(*input, "results.json"), &trials); err != nil {
		return err
	}
	if len(trials) != im.Total || im.Ticks < 1 {
		return fmt.Errorf("assay manifest/results mismatch")
	}
	wanted := map[string]bool{}
	for _, t := range trials {
		wanted[t.SourceHash] = true
	}
	sources := map[string]*world.World{}
	for _, row := range rows {
		if !wanted[row.StateSHA256] {
			continue
		}
		w, e := load(*source, row.Snapshot)
		if e != nil {
			return e
		}
		if kernel.Hash(w) != row.StateSHA256 || w.Tick != row.Tick || w.Config.Seed != row.Seed {
			return fmt.Errorf("source snapshot metadata mismatch")
		}
		if sources[row.StateSHA256] != nil {
			return fmt.Errorf("duplicate source snapshot")
		}
		sources[row.StateSHA256] = w
	}
	seen := map[string]bool{}
	for _, t := range trials {
		w := sources[t.SourceHash]
		key := fmt.Sprintf("%d/%s", t.Seed, t.Variant)
		if w == nil || seen[key] || w.Config.Seed != t.Seed || t.Summary.FromTick != w.Tick || t.Summary.ToTick-w.Tick != uint64(im.Ticks) {
			return fmt.Errorf("unmatched, duplicate or invalid replay trial")
		}
		seen[key] = true
		initial, e := experiment.CollectiveStart(w, t.Variant)
		if e != nil {
			return e
		}
		if kernel.Hash(initial) != t.InitialHash {
			return fmt.Errorf("assay initial hash mismatch")
		}
		final, e := load(*input, t.Snapshot)
		if e != nil {
			return e
		}
		if kernel.Hash(final) != t.FinalHash || final.Tick != t.Summary.ToTick || final.Config.Seed != t.Seed {
			return fmt.Errorf("assay final snapshot mismatch")
		}
	}
	if err := os.Mkdir(*dest, 0755); err != nil {
		return err
	}
	start := time.Now()
	status := map[string]any{"status": "running", "input": *input, "source": *source, "total": len(trials), "workers": *workers, "observation": c, "ticks": im.Ticks, "started_utc": start.UTC().Format(time.RFC3339)}
	if _, err := save(filepath.Join(*dest, "manifest.json"), status); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			status["status"] = "failed"
			status["error"] = err.Error()
			_, _ = save(filepath.Join(*dest, "manifest.json"), status)
		}
	}()
	old := runtime.GOMAXPROCS(*workers)
	defer runtime.GOMAXPROCS(old)
	type answer struct {
		r result
		e error
	}
	queue := make(chan experiment.EnvironmentTrial, len(trials))
	answers := make(chan answer, len(trials))
	for _, t := range trials {
		queue <- t
	}
	close(queue)
	var wg sync.WaitGroup
	for worker := 0; worker < min(*workers, len(trials)); worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range queue {
				r, e := discovery.Replay(sources[t.SourceHash], t.Variant, im.Ticks, c)
				v := result{}
				if e == nil && r.FinalHash != t.FinalHash {
					e = fmt.Errorf("replay diverged: seed %d %s", t.Seed, t.Variant)
				}
				if e == nil {
					v = compact(r)
					v.File = fmt.Sprintf("seed-%d-%s.json", t.Seed, t.Variant)
					v.Hash, e = save(filepath.Join(*dest, v.File), r)
				}
				answers <- answer{v, e}
			}
		}()
	}
	wg.Wait()
	close(answers)
	results := []result{}
	for a := range answers {
		if a.e != nil {
			if err == nil {
				err = a.e
			}
		} else {
			results = append(results, a.r)
		}
	}
	if err != nil {
		return err
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Seed != results[j].Seed {
			return results[i].Seed < results[j].Seed
		}
		return results[i].Treatment < results[j].Treatment
	})
	if _, err = save(filepath.Join(*dest, "index.json"), results); err != nil {
		return err
	}
	status["status"] = "complete"
	status["completed"] = len(results)
	status["seconds"] = time.Since(start).Seconds()
	if _, err = save(filepath.Join(*dest, "manifest.json"), status); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Verified %d replays in %.2f s: %s\n", len(results), status["seconds"], filepath.Join(*dest, "index.json"))
	return err
}
