// role-assay tests contributions of cells in previously preserved clonal groups.
package main

import (
	"crypto/sha256"
	"encoding/hex"
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
	"sync"
	"time"
)

type witness struct {
	Case      string                     `json:"case"`
	Seed      uint64                     `json:"seed"`
	Snapshot  string                     `json:"snapshot"`
	Hash      string                     `json:"snapshot_sha256"`
	Candidate experiment.ClonalCandidate `json:"candidate"`
}
type result struct {
	Witness         string `json:"witness"`
	Case            string `json:"case"`
	Seed            uint64 `json:"seed"`
	Snapshot        string `json:"released_snapshot"`
	ControlVerified bool   `json:"ordinary_control_verified"`
	experiment.RoleResult
}

func read(path string, dest any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, dest)
}
func save(path string, value any) error {
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
func main() {
	if e := run(os.Args[1:], os.Stdout); e != nil {
		fmt.Fprintln(os.Stderr, "role-assay:", e)
		os.Exit(1)
	}
}
func run(args []string, out io.Writer) (err error) {
	fs := flag.NewFlagSet("role-assay", flag.ContinueOnError)
	fs.SetOutput(out)
	input := fs.String("input", "data/multicell-mixed-witnesses", "completed witness directory")
	dest := fs.String("out", "", "new output directory")
	ticks := fs.Int("ticks", 5000, "ticks per matched continuation")
	every := fs.Int("every", 100, "sampling interval")
	workers := fs.Int("workers", 16, "parallel worlds")
	if e := fs.Parse(args); e != nil {
		if e == flag.ErrHelp {
			return nil
		}
		return e
	}
	if *dest == "" || *ticks < 1 || *every < 1 || *every > *ticks || *workers < 1 || *workers > 256 || fs.NArg() != 0 {
		return fmt.Errorf("provide out, ticks >= every > 0, workers 1..256")
	}
	var manifest struct {
		Status           string
		Total, Completed int
	}
	if e := read(filepath.Join(*input, "manifest.json"), &manifest); e != nil {
		return e
	}
	var witnesses []witness
	if e := read(filepath.Join(*input, "witnesses.json"), &witnesses); e != nil {
		return e
	}
	if manifest.Status != "complete" || len(witnesses) == 0 || manifest.Total != len(witnesses) || manifest.Completed != len(witnesses) {
		return fmt.Errorf("incomplete witness batch")
	}
	type job struct {
		index    int
		source   *world.World
		witness  witness
		protocol experiment.RoleProtocol
	}
	jobs := []job{}
	seen := map[string]bool{}
	for _, row := range witnesses {
		if row.Snapshot == "" || filepath.Base(row.Snapshot) != row.Snapshot || seen[row.Snapshot] {
			return fmt.Errorf("invalid or duplicate snapshot")
		}
		seen[row.Snapshot] = true
		f, e := os.Open(filepath.Join(*input, row.Snapshot))
		if e != nil {
			return e
		}
		w, e := kernel.Load(f)
		f.Close()
		if e != nil {
			return e
		}
		if kernel.Hash(w) != row.Hash || w.Tick != row.Candidate.Qualified || w.Config.Seed != row.Seed {
			return fmt.Errorf("witness metadata mismatch: %s", row.Snapshot)
		}
		if e := experiment.ValidateRoleMembers(w, row.Candidate.Members); e != nil {
			return e
		}
		actual := map[string]int{}
		for _, id := range row.Candidate.Members {
			actual[w.Particles[id].Genome]++
		}
		if len(actual) != len(row.Candidate.Genomes) || len(actual) < 2 {
			return fmt.Errorf("expected mixed-genome witness")
		}
		for _, g := range row.Candidate.Genomes {
			if actual[g.Hash] != g.Count {
				return fmt.Errorf("witness genome mismatch")
			}
			delete(actual, g.Hash)
		}
		for _, mode := range []string{"intact", "no-sharing", "no-peer-sharing", "no-signals", "no-bonds", "no-acquisition"} {
			donors := []uint64{0}
			if mode == "no-acquisition" {
				donors = row.Candidate.Members
			}
			for _, donor := range donors {
				jobs = append(jobs, job{len(jobs), w, row, experiment.RoleProtocol{Version: 1, Mode: mode, Donor: donor, Members: row.Candidate.Members, Ticks: *ticks, Every: *every}})
			}
		}
	}
	if e := os.Mkdir(*dest, 0755); e != nil {
		return e
	}
	start := time.Now()
	status := map[string]any{"status": "running", "input": *input, "workers": *workers, "total": len(jobs), "completed": 0, "protocol_version": 1, "snapshot_semantics": "Released physical states: resuming snapshots alone DOES NOT resume interventions. Replay from witnesses with recorded protocols."}
	if exe, e := os.Executable(); e == nil {
		if b, e := os.ReadFile(exe); e == nil {
			h := sha256.Sum256(b)
			status["binary_sha256"] = hex.EncodeToString(h[:])
		}
	}
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
	queue := make(chan job, len(jobs))
	for _, j := range jobs {
		queue <- j
	}
	close(queue)
	rows := make([]result, len(jobs))
	errors := make([]error, len(jobs))
	var wg sync.WaitGroup
	for i := 0; i < min(*workers, len(jobs)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				w, r, e := experiment.ContinueRoles(j.source, j.protocol)
				if e != nil {
					errors[j.index] = e
					continue
				}
				row := result{Witness: j.witness.Snapshot, Case: j.witness.Case, Seed: j.witness.Seed, RoleResult: r}
				if j.protocol.Mode == "intact" {
					f, openErr := os.Open(filepath.Join(*input, j.witness.Snapshot))
					if openErr != nil {
						errors[j.index] = openErr
						continue
					}
					control, loadErr := kernel.Load(f)
					f.Close()
					if loadErr != nil {
						errors[j.index] = loadErr
						continue
					}
					for n := 0; n < *ticks; n++ {
						kernel.Step(control)
					}
					if kernel.Hash(control) != r.FinalHash {
						errors[j.index] = fmt.Errorf("observer changed control")
						continue
					}
					row.ControlVerified = true
				}
				if kernel.Hash(j.source) != r.SourceHash {
					errors[j.index] = fmt.Errorf("source mutated")
					continue
				}
				row.Snapshot = fmt.Sprintf("r%03d-released.json", j.index+1)
				f, e := os.Create(filepath.Join(*dest, row.Snapshot))
				if e == nil {
					e = kernel.Save(f, w)
					ce := f.Close()
					if e == nil {
						e = ce
					}
				}
				rows[j.index] = row
				errors[j.index] = e
			}
		}()
	}
	wg.Wait()
	for _, e := range errors {
		if e != nil {
			return e
		}
	}
	if e := save(filepath.Join(*dest, "results.json"), rows); e != nil {
		return e
	}
	status["status"] = "complete"
	status["completed"] = len(rows)
	status["seconds"] = time.Since(start).Seconds()
	if e := save(filepath.Join(*dest, "manifest.json"), status); e != nil {
		return e
	}
	_, err = fmt.Fprintf(out, "Completed %d trials from %d witnesses in %.2f seconds: %s\n", len(rows), len(witnesses), status["seconds"], *dest)
	return err
}
