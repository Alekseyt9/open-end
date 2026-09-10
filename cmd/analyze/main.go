// Command analyze evaluates persistence from a completed batch's telemetry.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/observer"
	"os"
	"path/filepath"
)

type runSummary struct {
	Case         string
	Seed         uint64
	Tick         uint64
	Entities     int
	Genomes      int
	RecentCopies uint64
	RecentTicks  uint64
	StateSHA256  string
	Metrics      string
}

type runAnalysis struct {
	Case         string               `json:"case"`
	Seed         uint64               `json:"seed"`
	Tick         uint64               `json:"tick"`
	Entities     int                  `json:"entities"`
	Genomes      int                  `json:"genomes"`
	RecentCopies uint64               `json:"recent_copies"`
	RecentTicks  uint64               `json:"recent_ticks"`
	StateSHA256  string               `json:"state_sha256"`
	Persistence  observer.Persistence `json:"persistence"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "analyze:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	input := fs.String("input", "", "completed batch directory")
	windows := fs.Int("windows", 3, "consecutive reporting windows required")
	minCount := fs.Int("min-count", 5, "minimum live copies at each boundary")
	minCopies := fs.Uint64("min-copies", 5, "minimum offspring per window")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || fs.NArg() != 0 || *windows < 1 || *minCount < 1 || *minCopies < 1 {
		return fmt.Errorf("provide -input and positive thresholds")
	}
	manifestData, err := os.ReadFile(filepath.Join(*input, "manifest.json"))
	if err != nil {
		return err
	}
	var manifest struct {
		Status    string
		Completed int
		Total     int
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return err
	}
	if manifest.Status != "complete" || manifest.Completed != manifest.Total {
		return fmt.Errorf("batch is not complete")
	}
	data, err := os.ReadFile(filepath.Join(*input, "summary.json"))
	if err != nil {
		return err
	}
	var summaries []runSummary
	if err := json.Unmarshal(data, &summaries); err != nil {
		return err
	}
	if len(summaries) != manifest.Total || len(summaries) == 0 {
		return fmt.Errorf("summary does not match manifest")
	}
	results := make([]runAnalysis, 0, len(summaries))
	for _, s := range summaries {
		if s.Metrics == "" || s.Metrics != filepath.Base(s.Metrics) {
			return fmt.Errorf("invalid metrics filename")
		}
		frames, err := readFrames(filepath.Join(*input, s.Metrics), *windows+1)
		if err != nil {
			return err
		}
		p, err := observer.Persistent(frames, observer.PersistenceOptions{Windows: *windows, MinCount: *minCount, MinCopies: *minCopies})
		if err != nil {
			return fmt.Errorf("%s seed %d: %w", s.Case, s.Seed, err)
		}
		last := frames[len(frames)-1]
		if last.Tick != s.Tick || last.Entities != s.Entities || last.Genomes != s.Genomes {
			return fmt.Errorf("metrics and summary disagree")
		}
		results = append(results, runAnalysis{Case: s.Case, Seed: s.Seed, Tick: s.Tick, Entities: s.Entities, Genomes: s.Genomes, RecentCopies: s.RecentCopies, RecentTicks: s.RecentTicks, StateSHA256: s.StateSHA256, Persistence: p})
	}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func readFrames(path string, keep int) ([]observer.Metrics, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	frames := make([]observer.Metrics, 0, keep)
	d := json.NewDecoder(f)
	var lastTick uint64
	first := true
	for {
		var m observer.Metrics
		if err := d.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if !first && m.Tick <= lastTick {
			return nil, fmt.Errorf("non-increasing telemetry tick")
		}
		first = false
		lastTick = m.Tick
		if len(frames) == keep {
			copy(frames, frames[1:])
			frames = frames[:keep-1]
		}
		frames = append(frames, m)
	}
	return frames, nil
}
