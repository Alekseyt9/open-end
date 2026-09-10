package council

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"open-end/internal/dsl"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type TrialOptions struct {
	Ticks, Every, Workers int
	Window                uint64
}
type TrialRow struct {
	WorldID    string                 `json:"world_id"`
	Case       string                 `json:"case"`
	Seed       uint64                 `json:"seed"`
	Variant    string                 `json:"variant"`
	SourceHash string                 `json:"source_snapshot_sha256"`
	FinalHash  string                 `json:"final_snapshot_sha256"`
	RulesHash  string                 `json:"rules_sha256"`
	Snapshot   string                 `json:"snapshot"`
	Metrics    string                 `json:"metrics"`
	Summary    observer.WindowSummary `json:"summary"`
	Dynamics   observer.Dynamics      `json:"dynamics"`
	Error      string                 `json:"error,omitempty"`
}
type trialManifest struct {
	Status       string       `json:"status"`
	RequestID    string       `json:"request_sha256"`
	ResponseHash string       `json:"response_sha256"`
	Kernel       string       `json:"kernel_version"`
	Options      TrialOptions `json:"options"`
	Total        int          `json:"total"`
	Completed    int          `json:"completed"`
	Seconds      float64      `json:"seconds"`
}
type trialJob struct {
	World   WorldBrief
	Variant string
	Module  *dsl.Module
}

// Trial runs paired continuations from immutable snapshots, never accepting
// shell commands or genotype edits. Results are sorted independently of worker
// completion order. A failed run leaves an explicitly incomplete manifest.
func Trial(round, responsePath, out string, opts TrialOptions, progress io.Writer) ([]TrialRow, error) {
	if opts.Ticks < 1 || opts.Every < 1 || opts.Workers < 1 || opts.Workers > 256 || opts.Window == 0 {
		return nil, fmt.Errorf("invalid trial options")
	}
	c, err := Check(round, responsePath)
	if err != nil {
		return nil, err
	}
	if len(c.Modules) == 0 {
		return nil, fmt.Errorf("no proposals to trial; observer-only response is valid for check")
	}
	return executeTrial(c, round, out, opts, progress)
}

func executeTrial(c Checked, round, out string, opts TrialOptions, progress io.Writer) ([]TrialRow, error) {
	jobs := []trialJob{}
	for _, w := range c.Request.Worlds {
		if uint64(opts.Ticks) > ^uint64(0)-w.Tick {
			return nil, fmt.Errorf("tick overflow")
		}
		jobs = append(jobs, trialJob{World: w, Variant: "control"})
		for _, p := range c.Response.Proposals {
			jobs = append(jobs, trialJob{World: w, Variant: p.ID, Module: c.Modules[p.ID]})
		}
	}
	if err := os.Mkdir(out, 0755); err != nil {
		return nil, err
	}
	manifest := trialManifest{Status: "running", RequestID: c.Request.ID, ResponseHash: c.ResponseHash, Kernel: kernel.Version, Options: opts, Total: len(jobs)}
	if err := writeJSON(filepath.Join(out, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	// Freeze the exact accepted reply and request alongside the run.
	if c.ResponseHash != "" {
		if err := writeJSON(filepath.Join(out, "response.json"), c.Response); err != nil {
			return nil, err
		}
	}
	if err := writeJSON(filepath.Join(out, "request.json"), c.Request); err != nil {
		return nil, err
	}
	for id, m := range c.Modules {
		if err := writeJSON(filepath.Join(out, id+".rules.json"), m.Source); err != nil {
			return nil, err
		}
	}
	oldProcs := runtime.GOMAXPROCS(opts.Workers)
	defer runtime.GOMAXPROCS(oldProcs)
	start := time.Now()
	queue := make(chan trialJob, len(jobs))
	done := make(chan TrialRow, len(jobs))
	for _, job := range jobs {
		queue <- job
	}
	close(queue)
	var wg sync.WaitGroup
	for i := 0; i < min(opts.Workers, len(jobs)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range queue {
				done <- runBranch(round, out, job, opts)
			}
		}()
	}
	go func() { wg.Wait(); close(done) }()
	rows := []TrialRow{}
	var runErr error
	for row := range done {
		rows = append(rows, row)
		if row.Error != "" {
			runErr = fmt.Errorf("one or more trial branches failed")
		} else {
			manifest.Completed++
		}
		if _, err := fmt.Fprintf(progress, "completed %d/%d: %s %s %s\n", len(rows), len(jobs), row.WorldID, row.Variant, row.Dynamics.Status); err != nil {
			runErr = err
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].WorldID != rows[j].WorldID {
			return rows[i].WorldID < rows[j].WorldID
		}
		return rows[i].Variant < rows[j].Variant
	})
	manifest.Seconds = time.Since(start).Seconds()
	manifest.Status = "failed"
	if runErr == nil {
		manifest.Status = "complete"
	}
	if err := writeJSON(filepath.Join(out, "results.json"), rows); err != nil {
		return rows, err
	}
	if err := writeNew(filepath.Join(out, "comparison.md"), []byte(compare(rows))); err != nil {
		return rows, err
	}
	// Only this invocation created manifest.json. Atomic replacement publishes
	// completion after all referenced results have been written.
	b, _ := json.MarshalIndent(manifest, "", "  ")
	if err := writeNew(filepath.Join(out, "manifest.tmp"), append(b, '\n')); err != nil {
		return rows, err
	}
	if err := os.Rename(filepath.Join(out, "manifest.tmp"), filepath.Join(out, "manifest.json")); err != nil {
		return rows, err
	}
	return rows, runErr
}
func runBranch(round, out string, job trialJob, opts TrialOptions) (r TrialRow) {
	r = TrialRow{WorldID: job.World.ID, Case: job.World.Case, Seed: job.World.Seed, Variant: job.Variant, SourceHash: job.World.SnapshotHash}
	stem := job.World.ID + "-" + job.Variant
	r.Snapshot = stem + ".snapshot.json"
	r.Metrics = stem + ".jsonl"
	fail := func(err error) TrialRow { r.Error = err.Error(); return r }
	w, err := loadWorld(filepath.Join(round, job.World.ID+".snapshot.json"))
	if err != nil {
		return fail(err)
	}
	if kernel.Hash(w) != r.SourceHash {
		return fail(fmt.Errorf("source changed after validation"))
	}
	if w.RuleState != nil && len(w.RuleState.Pending) > 0 {
		return fail(fmt.Errorf("pending rule changes"))
	}
	tracker := observer.NewTracker(w)
	frames := []observer.Metrics{tracker.Frame(w)}
	if job.Module != nil {
		if err := kernel.ReloadRules(w, job.Module); err != nil {
			return fail(err)
		}
	}
	r.RulesHash = job.World.Rules.Hash
	if job.Module != nil {
		r.RulesHash = job.Module.Hash
	}
	f, err := os.OpenFile(filepath.Join(out, r.Metrics), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fail(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(frames[0]); err != nil {
		return fail(err)
	}
	for i := 1; i <= opts.Ticks; i++ {
		kernel.StepObserved(w, tracker)
		if i%opts.Every == 0 || i == opts.Ticks {
			m := tracker.Frame(w)
			if !m.Telemetry.Complete {
				return fail(fmt.Errorf("incomplete telemetry"))
			}
			if err := enc.Encode(m); err != nil {
				return fail(err)
			}
			frames = append(frames, m)
			cut := uint64(0)
			if m.Tick > opts.Window {
				cut = m.Tick - opts.Window
			}
			for len(frames) > 2 && frames[1].Tick <= cut {
				frames[0] = observer.Metrics{}
				frames = frames[1:]
			}
		}
	}
	if err := f.Close(); err != nil {
		return fail(err)
	}
	var snapshot bytes.Buffer
	if err := kernel.Save(&snapshot, w); err != nil {
		return fail(err)
	}
	if err := writeNew(filepath.Join(out, r.Snapshot), snapshot.Bytes()); err != nil {
		return fail(err)
	}
	r.FinalHash = kernel.Hash(w)
	r.Summary, err = observer.Summarize(frames, opts.Window)
	if err != nil {
		return fail(err)
	}
	r.Dynamics, err = observer.DetectDynamics(frames, opts.Window, observer.DefaultDynamicsConfig())
	if err != nil {
		return fail(err)
	}
	return r
}
func compare(rows []TrialRow) string {
	var b strings.Builder
	b.WriteString("# Пробное сравнение правил\n\nКонтроль и предложение продолжают один исходный snapshot с одинаковым начальным RNG. После расхождения состояний одинаковый RNG не означает одинаковые события. Автоматического победителя и изменения исходных миров нет.\n\n| Мир | Вариант | Популяция | Эффективное разнообразие | Крупнейшая структура | Копии за окно | Статус |\n|---|---|---:|---:|---:|---:|---|\n")
	for _, r := range rows {
		if r.Error != "" {
			fmt.Fprintf(&b, "| %s | %s | — | — | — | — | ошибка |\n", r.WorldID, r.Variant)
			continue
		}
		fmt.Fprintf(&b, "| %s (%s, seed %d) | %s | %d | %.3f | %d | %d | %s |\n", r.WorldID, r.Case, r.Seed, r.Variant, r.Summary.Population.End, r.Summary.DiversityEnd.EffectiveGenomes, r.Summary.StructuresEnd.Largest, r.Summary.Copies, r.Dynamics.Status)
	}
	b.WriteString("\nЧисла относятся к последнему запрошенному окну; его реальные границы сохранены в `results.json`. Различия требуют интерпретации, повторов и более длинных прогонов. Поведенческие хеши разных физических правил напрямую не ранжируются.\n")
	return b.String()
}
