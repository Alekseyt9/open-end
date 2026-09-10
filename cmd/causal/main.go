// causal tests preselected group descriptions with seed-held-out forecasting and local interventions.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/causal"
	"open-end/internal/discovery"
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
		fmt.Fprintln(os.Stderr, "causal:", err)
		os.Exit(1)
	}
}
func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func read(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func save(path string, v any) (string, error) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return "", e
	}
	b = append(b, '\n')
	if e = os.WriteFile(path, b, 0644); e != nil {
		return "", e
	}
	return digest(b), nil
}
func safe(dir, name string) (string, error) {
	if name == "" || filepath.Base(name) != name {
		return "", fmt.Errorf("invalid artifact filename")
	}
	return filepath.Join(dir, name), nil
}
func complete(dir string) (int, error) {
	var m struct {
		Status           string
		Total, Completed int
	}
	e := read(filepath.Join(dir, "manifest.json"), &m)
	if e == nil && (m.Status != "complete" || m.Total < 1 || m.Total != m.Completed) {
		e = fmt.Errorf("incomplete input batch")
	}
	return m.Total, e
}

type source struct {
	w          *world.World
	groups     []causal.Group
	reportHash string
}
type artifact struct {
	File string `json:"file"`
	Hash string `json:"sha256"`
}
type pairResult struct {
	Seed          uint64              `json:"seed"`
	Group         string              `json:"group"`
	Control       causal.Outcome      `json:"control"`
	Local         causal.Intervention `json:"local_cut"`
	Outside       causal.Intervention `json:"outside_cut"`
	SurvivalDelta float64             `json:"local_minus_control_survival"`
	OutsideDelta  *float64            `json:"outside_minus_control_survival"`
	SpecificDelta *float64            `json:"local_minus_outside_survival"`
}

func run(args []string, out io.Writer) (err error) {
	c := causal.DefaultConfig()
	fs := flag.NewFlagSet("causal", flag.ContinueOnError)
	fs.SetOutput(out)
	input := fs.String("input", "", "completed discovery directory")
	snapshots := fs.String("snapshots", "", "completed group assay containing discovery final snapshots")
	dest := fs.String("out", "", "new output directory")
	workers := fs.Int("workers", 16, "independent continuation workers")
	fs.IntVar(&c.Horizon, "horizon", c.Horizon, "prediction and intervention horizon")
	fs.IntVar(&c.Blocks, "blocks", c.Blocks, "prospective forecast intervals")
	fs.IntVar(&c.Permutations, "permutations", c.Permutations, "matched random partitions")
	fs.Float64Var(&c.Ridge, "ridge", c.Ridge, "fixed ridge coefficient, never tuned on evaluation seeds")
	if e := fs.Parse(args); e != nil {
		if e == flag.ErrHelp {
			return nil
		}
		return e
	}
	if *input == "" || *snapshots == "" || *dest == "" || *workers < 1 || *workers > 256 || fs.NArg() != 0 {
		return fmt.Errorf("provide input/snapshots/out and workers 1..256")
	}
	if e := c.Validate(); e != nil {
		return e
	}
	total, e := complete(*input)
	if e != nil {
		return e
	}
	snapshotTotal, e := complete(*snapshots)
	if e != nil {
		return e
	}
	var index []struct {
		Seed      uint64 `json:"seed"`
		Treatment string `json:"treatment"`
		File      string `json:"report"`
		Hash      string `json:"report_sha256"`
		Final     string `json:"verified_final_sha256"`
	}
	if e := read(filepath.Join(*input, "index.json"), &index); e != nil {
		return e
	}
	if len(index) != total {
		return fmt.Errorf("discovery manifest/index mismatch")
	}
	var rows []struct {
		Snapshot, StateSHA256 string
		Seed, Tick            uint64
	}
	if e := read(filepath.Join(*snapshots, "summary.json"), &rows); e != nil {
		return e
	}
	if len(rows) != snapshotTotal {
		return fmt.Errorf("snapshot manifest/summary mismatch")
	}
	byHash := map[string]string{}
	for _, r := range rows {
		if _, ok := byHash[r.StateSHA256]; ok {
			return fmt.Errorf("duplicate snapshot hash")
		}
		byHash[r.StateSHA256] = r.Snapshot
	}
	sources := []source{}
	seen := map[uint64]bool{}
	for _, entry := range index {
		if entry.Treatment != "intact" {
			continue
		}
		if seen[entry.Seed] {
			return fmt.Errorf("duplicate seed")
		}
		seen[entry.Seed] = true
		path, e := safe(*input, entry.File)
		if e != nil {
			return e
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if digest(b) != entry.Hash {
			return fmt.Errorf("discovery report hash mismatch")
		}
		var r discovery.Report
		if e := json.Unmarshal(b, &r); e != nil {
			return e
		}
		if r.Seed != entry.Seed || r.FinalHash != entry.Final || r.Treatment != entry.Treatment {
			return fmt.Errorf("discovery metadata mismatch")
		}
		path, e = safe(*snapshots, byHash[entry.Final])
		if e != nil {
			return e
		}
		f, e := os.Open(path)
		if e != nil {
			return e
		}
		w, e := kernel.Load(f)
		f.Close()
		if e != nil {
			return e
		}
		if w.Config.Seed != entry.Seed {
			return fmt.Errorf("snapshot seed mismatch")
		}
		groups, e := causal.Select(w, r, c)
		if e != nil {
			return e
		}
		sources = append(sources, source{w, groups, entry.Hash})
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].w.Config.Seed < sources[j].w.Config.Seed })
	eligible := 0
	for _, s := range sources {
		if len(s.groups) > 0 {
			eligible++
		}
	}
	if eligible < 3 {
		return fmt.Errorf("need at least three independent seeds with selected groups")
	}
	if e := os.Mkdir(*dest, 0755); e != nil {
		return e
	}
	start := time.Now()
	type job struct {
		s    source
		g    causal.Group
		mode string
	}
	jobs := []job{}
	for _, s := range sources {
		jobs = append(jobs, job{s: s, mode: "baseline"})
		for _, g := range s.groups {
			if g.Partition == "bonds" {
				jobs = append(jobs, job{s, g, "local-cut"}, job{s, g, "outside-cut"})
			}
		}
	}
	status := map[string]any{"status": "running", "input": *input, "snapshots": *snapshots, "configuration": c, "workers": *workers, "total": len(jobs), "source_worlds": len(sources), "eligible_seeds": eligible, "started_utc": start.UTC().Format(time.RFC3339)}
	if _, e := save(filepath.Join(*dest, "manifest.json"), status); e != nil {
		return e
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
		base         *causal.Baseline
		intervention *causal.Intervention
		file         artifact
		err          error
	}
	queue := make(chan job, len(jobs))
	answers := make(chan answer, len(jobs))
	for _, j := range jobs {
		queue <- j
	}
	close(queue)
	var wg sync.WaitGroup
	for worker := 0; worker < min(*workers, len(jobs)); worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				a := answer{}
				if j.mode == "baseline" {
					r, e := causal.Forecast(j.s.w, j.s.groups, c)
					a.base = &r
					a.err = e
					a.file.File = fmt.Sprintf("seed-%d-baseline.json", j.s.w.Config.Seed)
				} else {
					r, e := causal.Intervene(j.s.w, j.g, j.mode, c.Horizon)
					a.intervention = &r
					a.err = e
					a.file.File = fmt.Sprintf("seed-%d-%s-%s.json", j.s.w.Config.Seed, j.g.ID, j.mode)
				}
				if a.err == nil {
					var value any = a.base
					if a.intervention != nil {
						value = a.intervention
					}
					a.file.Hash, a.err = save(filepath.Join(*dest, a.file.File), value)
				}
				answers <- a
			}
		}()
	}
	wg.Wait()
	close(answers)
	bases := map[uint64]causal.Baseline{}
	interventions := map[string]causal.Intervention{}
	files := []artifact{}
	for a := range answers {
		if a.err != nil {
			if err == nil {
				err = a.err
			}
			continue
		}
		files = append(files, a.file)
		if a.base != nil {
			bases[a.base.Seed] = *a.base
		} else {
			v := *a.intervention
			interventions[fmt.Sprintf("%d/%s/%s", v.Seed, v.Group, v.Mode)] = v
		}
	}
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].File < files[j].File })
	samples := []causal.Sample{}
	pairs := []pairResult{}
	provenance := []map[string]any{}
	for _, s := range sources {
		seed := s.w.Config.Seed
		b := bases[seed]
		samples = append(samples, b.Samples...)
		provenance = append(provenance, map[string]any{"seed": seed, "source_sha256": b.SourceHash, "discovery_report_sha256": s.reportHash, "control_horizon_sha256": b.FirstHash, "control_final_sha256": b.FinalHash})
		for _, g := range s.groups {
			if g.Partition != "bonds" {
				continue
			}
			p := pairResult{Seed: seed, Group: g.ID, Control: b.Control[g.ID], Local: interventions[fmt.Sprintf("%d/%s/local-cut", seed, g.ID)], Outside: interventions[fmt.Sprintf("%d/%s/outside-cut", seed, g.ID)]}
			p.SurvivalDelta = float64(p.Local.Outcome.Alive-p.Control.Alive) / float64(len(g.Members))
			if p.Outside.Available {
				x := float64(p.Outside.Outcome.Alive-p.Control.Alive) / float64(len(g.Members))
				y := p.SurvivalDelta - x
				p.OutsideDelta = &x
				p.SpecificDelta = &y
			}
			pairs = append(pairs, p)
		}
	}
	evaluation, e := causal.Evaluate(samples, c.Ridge)
	if e != nil {
		return e
	}
	hash, err := save(filepath.Join(*dest, "prediction.json"), evaluation)
	if err != nil {
		return err
	}
	files = append(files, artifact{"prediction.json", hash})
	hash, err = save(filepath.Join(*dest, "interventions.json"), pairs)
	if err != nil {
		return err
	}
	files = append(files, artifact{"interventions.json", hash})
	sort.Slice(files, func(i, j int) bool { return files[i].File < files[j].File })
	if _, err = save(filepath.Join(*dest, "index.json"), map[string]any{"artifacts": files, "sources": provenance, "micro_features": causal.MicroFeatures, "macro_features": causal.MacroFeatures}); err != nil {
		return err
	}
	status["status"] = "complete"
	status["completed"] = len(jobs)
	status["selected_groups"] = len(pairs)
	status["forecast_samples"] = len(samples)
	status["seconds"] = time.Since(start).Seconds()
	if _, err = save(filepath.Join(*dest, "manifest.json"), status); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Analyzed %d groups from %d eligible seeds in %.2f s: %s\n", len(pairs), eligible, status["seconds"], *dest)
	return err
}
