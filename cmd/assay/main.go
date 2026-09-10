package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Condition   string  `json:"condition"`
	Seed        uint64  `json:"seed"`
	Ticks       int     `json:"ticks"`
	InitialA    int     `json:"initial_a"`
	InitialB    int     `json:"initial_b"`
	FinalA      int     `json:"final_a"`
	FinalB      int     `json:"final_b"`
	CopiesA     uint64  `json:"copies_a"`
	CopiesB     uint64  `json:"copies_b"`
	RecentA     uint64  `json:"recent_copies_a"`
	RecentB     uint64  `json:"recent_copies_b"`
	RecentTicks int     `json:"recent_ticks"`
	StateHash   string  `json:"state_sha256"`
	Seconds     float64 `json:"seconds"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "assay:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("assay", flag.ContinueOnError)
	extract := fs.String("extract", "", "extract pair from snapshot")
	a := fs.String("a", "", "source genome A hash")
	b := fs.String("b", "", "source genome B hash")
	pairPath := fs.String("pair", "", "portable pair JSON")
	output := fs.String("output", "", "new output directory")
	workers := fs.Int("workers", 16, "independent single-thread worker processes")
	seeds := fs.Int("seeds", 16, "use placement seeds 1..N")
	ticks := fs.Int("ticks", 100000, "ticks per assay")
	every := fs.Int("every", 10000, "report interval and final rate window")
	founders := fs.Int("founders", 32, "even founder count for full treatments")
	condition := fs.String("worker", "", "internal single-condition worker")
	seed := fs.Uint64("seed", 1, "worker placement seed")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *extract != "" {
		f, err := os.Open(*extract)
		if err != nil {
			return err
		}
		defer f.Close()
		w, err := kernel.Load(f)
		if err != nil {
			return err
		}
		p, err := experiment.Extract(w, *a, *b)
		if err != nil {
			return err
		}
		return json.NewEncoder(out).Encode(p)
	}
	if *ticks < 1 || *every < 1 || *ticks%*every != 0 || *workers < 1 || *workers > 256 || *seeds < 1 || *seeds > 10000 || *founders < 2 || *founders > 1024 || *founders%2 != 0 {
		return fmt.Errorf("positive limits required; ticks must be divisible by every")
	}
	f, err := os.Open(*pairPath)
	if err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	if stat.Size() > 1<<20 {
		f.Close()
		return fmt.Errorf("pair exceeds 1 MiB")
	}
	var pair experiment.Pair
	d := json.NewDecoder(io.LimitReader(f, 1<<20))
	d.DisallowUnknownFields()
	err = d.Decode(&pair)
	if err == nil {
		var extra any
		if d.Decode(&extra) != io.EOF {
			err = fmt.Errorf("pair must contain one JSON object")
		}
	}
	f.Close()
	if err != nil {
		return err
	}
	if err := pair.Validate(); err != nil {
		return err
	}
	if *condition != "" {
		if *output == "" {
			return fmt.Errorf("worker requires output stem")
		}
		r, err := single(pair, *seed, *condition, *founders, *ticks, *every, *output)
		if err != nil {
			return err
		}
		return json.NewEncoder(out).Encode(r)
	}
	if *output == "" {
		return fmt.Errorf("provide -output")
	}
	if err := os.Mkdir(*output, 0755); err != nil {
		return err
	}
	pairAbs, err := filepath.Abs(*pairPath)
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	start := time.Now()
	manifest := map[string]any{"status": "running", "workers": *workers, "seeds": *seeds, "ticks": *ticks, "every": *every, "founders": *founders, "pair": pair, "kernel": kernel.Version, "rules": kernel.RuleVersion}
	saveManifest := func() error { return writeJSON(filepath.Join(*output, "manifest.json"), manifest) }
	if err := saveManifest(); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type job struct {
		condition string
		seed      int
	}
	type response struct {
		result Result
		err    error
	}
	jobs := make(chan job, len(experiment.AssayCases)**seeds)
	responses := make(chan response, *workers)
	for _, c := range experiment.AssayCases {
		for s := 1; s <= *seeds; s++ {
			jobs <- job{c, s}
		}
	}
	close(jobs)
	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if ctx.Err() != nil {
					return
				}
				stem := filepath.Join(*output, fmt.Sprintf("%s-seed%d", j.condition, j.seed))
				cmd := exec.CommandContext(ctx, exe, "-worker", j.condition, "-seed", strconv.Itoa(j.seed), "-pair", pairAbs, "-output", stem, "-ticks", strconv.Itoa(*ticks), "-every", strconv.Itoa(*every), "-founders", strconv.Itoa(*founders))
				configureProcess(cmd)
				for _, entry := range os.Environ() {
					if !strings.HasPrefix(strings.ToUpper(entry), "GOMAXPROCS=") {
						cmd.Env = append(cmd.Env, entry)
					}
				}
				cmd.Env = append(cmd.Env, "GOMAXPROCS=1")
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err := cmd.Run()
				var r Result
				if err == nil {
					err = json.Unmarshal(stdout.Bytes(), &r)
				} else {
					err = fmt.Errorf("%s seed %d: %w: %s", j.condition, j.seed, err, stderr.String())
				}
				responses <- response{r, err}
			}
		}()
	}
	go func() { wg.Wait(); close(responses) }()
	results := []Result{}
	var failure error
	for response := range responses {
		if response.err != nil {
			if failure == nil {
				failure = response.err
			}
			cancel()
			continue
		}
		results = append(results, response.result)
		fmt.Fprintf(out, "completed %d/%d: %s seed=%d A=%d B=%d\n", len(results), len(experiment.AssayCases)**seeds, response.result.Condition, response.result.Seed, response.result.FinalA, response.result.FinalB)
	}
	if failure == nil && ctx.Err() != nil {
		failure = ctx.Err()
	}
	manifest["seconds"] = time.Since(start).Seconds()
	manifest["completed"] = len(results)
	manifest["status"] = "complete"
	if failure != nil {
		manifest["status"] = "failed"
		manifest["error"] = failure.Error()
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Condition != results[j].Condition {
			return results[i].Condition < results[j].Condition
		}
		return results[i].Seed < results[j].Seed
	})
	if err := writeJSON(filepath.Join(*output, "summary.json"), results); err != nil {
		return err
	}
	if err := saveManifest(); err != nil {
		return err
	}
	return failure
}

func single(pair experiment.Pair, seed uint64, condition string, founders, ticks, every int, stem string) (Result, error) {
	start := time.Now()
	w, err := experiment.Inoculate(pair, seed, condition, founders)
	if err != nil {
		return Result{}, err
	}
	r := Result{Condition: condition, Seed: seed, Ticks: ticks, RecentTicks: every}
	count := func() (a, b int) {
		for _, p := range w.Particles {
			if p.Genome == pair.A.Hash {
				a++
			}
			if p.Genome == pair.B.Hash {
				b++
			}
		}
		return
	}
	copies := func(hash string) uint64 {
		if g := w.Genomes[hash]; g != nil {
			return g.Copies
		}
		return 0
	}
	r.InitialA, r.InitialB = count()
	f, err := os.OpenFile(stem+".jsonl", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return r, err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	tracker := observer.NewTracker(w)
	if err := enc.Encode(tracker.Frame(w)); err != nil {
		return r, err
	}
	var beforeA, beforeB uint64
	for t := 0; t < ticks; t++ {
		if t == ticks-every {
			beforeA, beforeB = copies(pair.A.Hash), copies(pair.B.Hash)
		}
		kernel.StepObserved(w, tracker)
		if (t+1)%every == 0 {
			if err := enc.Encode(tracker.Frame(w)); err != nil {
				return r, err
			}
		}
	}
	if err := w.Validate(); err != nil {
		return r, err
	}
	r.FinalA, r.FinalB = count()
	r.CopiesA, r.CopiesB = copies(pair.A.Hash), copies(pair.B.Hash)
	r.RecentA = r.CopiesA - beforeA
	r.RecentB = r.CopiesB - beforeB
	r.StateHash = kernel.Hash(w)
	r.Seconds = time.Since(start).Seconds()
	return r, nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}
