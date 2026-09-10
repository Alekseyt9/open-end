// summarize explains a recorded tick window from exact event telemetry.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/observer"
	"os"
	"path/filepath"
	"sort"
)

type result struct {
	Case     string                 `json:"case,omitempty"`
	Seed     uint64                 `json:"seed,omitempty"`
	Metrics  string                 `json:"metrics"`
	Summary  observer.WindowSummary `json:"summary"`
	Dynamics *observer.Dynamics     `json:"dynamics,omitempty"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "summarize:", err)
		os.Exit(1)
	}
}
func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("summarize", flag.ContinueOnError)
	fs.SetOutput(out)
	input := fs.String("input", "", "JSONL metrics file or completed experiment directory")
	window := fs.Uint64("window", 10000, "requested number of recent ticks; uses whole recorded intervals")
	format := fs.String("format", "text", "text or json")
	detect := fs.Bool("detect", false, "include novelty and stagnation heuristics")
	detectionConfig := observer.DefaultDynamicsConfig()
	fs.Uint64Var(&detectionConfig.MinTicks, "detect-min-ticks", detectionConfig.MinTicks, "minimum observed duration for dynamics detection")
	fs.Float64Var(&detectionConfig.PlateauTolerance, "detect-tolerance", detectionConfig.PlateauTolerance, "relative plateau and structure distribution tolerance (0,1]")
	fs.IntVar(&detectionConfig.BehaviorResolution, "detect-resolution", detectionConfig.BehaviorResolution, "behavior quantization bins per octave (1..64)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if *input == "" || *window == 0 || fs.NArg() != 0 || (*format != "text" && *format != "json") {
		return fmt.Errorf("provide -input, a positive -window and -format text|json")
	}
	info, err := os.Stat(*input)
	if err != nil {
		return err
	}
	results := []result{}
	analyze := func(path string) (observer.WindowSummary, *observer.Dynamics, error) {
		f, err := os.Open(path)
		if err != nil {
			return observer.WindowSummary{}, nil, err
		}
		defer f.Close()
		frames, err := readWindow(f, *window)
		if err != nil {
			return observer.WindowSummary{}, nil, err
		}
		summary, err := observer.Summarize(frames, *window)
		if err != nil || !*detect {
			return summary, nil, err
		}
		dynamics, err := observer.DetectDynamics(frames, *window, detectionConfig)
		return summary, &dynamics, err
	}
	if info.IsDir() {
		var manifest struct {
			Status           string
			Completed, Total int
		}
		if err := readJSON(filepath.Join(*input, "manifest.json"), &manifest); err != nil {
			return err
		}
		if manifest.Status != "complete" || manifest.Completed != manifest.Total || manifest.Total < 1 {
			return fmt.Errorf("batch is not complete")
		}
		var rows []struct {
			Case              string
			Seed, Tick        uint64
			Entities, Genomes int
			Metrics           string
		}
		if err := readJSON(filepath.Join(*input, "summary.json"), &rows); err != nil {
			return err
		}
		if len(rows) != manifest.Total {
			return fmt.Errorf("manifest and summary disagree")
		}
		seen := map[string]bool{}
		for _, row := range rows {
			if row.Metrics == "" || row.Metrics != filepath.Base(row.Metrics) || seen[row.Metrics] {
				return fmt.Errorf("invalid or duplicate metrics filename")
			}
			seen[row.Metrics] = true
			r, dynamics, err := analyze(filepath.Join(*input, row.Metrics))
			if err != nil {
				return fmt.Errorf("%s seed %d: %w", row.Case, row.Seed, err)
			}
			if r.ToTick != row.Tick || r.Population.End != int64(row.Entities) || r.Genomes.End != int64(row.Genomes) || r.Seed != row.Seed {
				return fmt.Errorf("metrics and batch summary disagree")
			}
			results = append(results, result{Case: row.Case, Seed: row.Seed, Metrics: row.Metrics, Summary: r, Dynamics: dynamics})
		}
		sort.Slice(results, func(i, j int) bool {
			a, b := results[i], results[j]
			if a.Case != b.Case {
				return a.Case < b.Case
			}
			return a.Seed < b.Seed
		})
	} else {
		r, dynamics, err := analyze(*input)
		if err != nil {
			return err
		}
		results = append(results, result{Seed: r.Seed, Metrics: filepath.Base(*input), Summary: r, Dynamics: dynamics})
	}
	if *format == "json" {
		e := json.NewEncoder(out)
		e.SetIndent("", "  ")
		return e.Encode(results)
	}
	for i, r := range results {
		if i > 0 {
			if _, err := fmt.Fprintln(out); err != nil {
				return err
			}
		}
		label := r.Case
		if label == "" {
			label = "Мир"
		}
		if _, err := fmt.Fprintf(out, "%s (seed %d) — %s\n", label, r.Seed, r.Metrics); err != nil {
			return err
		}
		for _, line := range r.Summary.Narrative {
			if _, err := fmt.Fprintln(out, line); err != nil {
				return err
			}
		}
		if r.Dynamics != nil {
			if _, err := fmt.Fprintf(out, "Детектор: %s\n", r.Dynamics.Status); err != nil {
				return err
			}
			for _, line := range r.Dynamics.Reasons {
				if _, err := fmt.Fprintln(out, line); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func readJSON(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)
	if err := d.Decode(v); err != nil {
		return err
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fmt.Errorf("expected one JSON document: %s", path)
	}
	return nil
}
func summarizeFile(path string, window uint64) (observer.WindowSummary, error) {
	f, err := os.Open(path)
	if err != nil {
		return observer.WindowSummary{}, err
	}
	defer f.Close()
	frames, err := readWindow(f, window)
	if err != nil {
		return observer.WindowSummary{}, err
	}
	return observer.Summarize(frames, window)
}
func readWindow(in io.Reader, window uint64) ([]observer.Metrics, error) {
	if window == 0 {
		return nil, fmt.Errorf("window must be positive")
	}
	d := json.NewDecoder(in)
	frames := []observer.Metrics{}
	for {
		var m observer.Metrics
		if err := d.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(frames) > 0 && m.Tick <= frames[len(frames)-1].Tick {
			return nil, fmt.Errorf("non-increasing telemetry tick")
		}
		frames = append(frames, m)
		cut := uint64(0)
		if m.Tick > window {
			cut = m.Tick - window
		}
		for len(frames) > 2 && frames[1].Tick <= cut {
			frames[0] = observer.Metrics{}
			frames = frames[1:]
		}
	}
	return frames, nil
}
