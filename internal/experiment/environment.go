package experiment

import (
	"bytes"
	"fmt"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
)

type EnvironmentTrial struct {
	Case        string                 `json:"case"`
	Seed        uint64                 `json:"seed"`
	Variant     string                 `json:"variant"`
	SourceHash  string                 `json:"source_sha256"`
	InitialHash string                 `json:"initial_sha256"`
	FinalHash   string                 `json:"final_sha256"`
	Snapshot    string                 `json:"snapshot"`
	Metrics     string                 `json:"metrics"`
	Summary     observer.WindowSummary `json:"summary"`
}

// ContinueEnvironment changes only the terrain feedback switch. It preserves
// every particle, field, inherited policy, RNG state and cumulative ledger.
// The original remains immutable. This intervention is outside world physics.
func ContinueEnvironment(source *world.World, mode string, ticks, every int) (*world.World, []observer.Metrics, EnvironmentTrial, error) {
	r := EnvironmentTrial{Variant: mode}
	if source == nil || source.Config.Environment == "" || ticks < 1 || every < 1 || (mode != "coupled" && mode != "inert") {
		return nil, nil, r, fmt.Errorf("invalid environment continuation")
	}
	var buf bytes.Buffer
	if err := kernel.Save(&buf, source); err != nil {
		return nil, nil, r, err
	}
	r.SourceHash = kernel.Hash(source)
	r.Seed = source.Config.Seed
	w, err := kernel.Load(&buf)
	if err != nil {
		return nil, nil, r, err
	}
	w.Config.Environment = mode
	r.InitialHash = kernel.Hash(w)
	tracker := observer.NewTracker(w)
	frames := []observer.Metrics{tracker.Frame(w)}
	for i := 1; i <= ticks; i++ {
		kernel.StepObserved(w, tracker)
		if i%every == 0 || i == ticks {
			frames = append(frames, tracker.Frame(w))
		}
	}
	if err := w.Validate(); err != nil {
		return nil, nil, r, err
	}
	r.Summary, err = observer.Summarize(frames, uint64(ticks))
	r.FinalHash = kernel.Hash(w)
	return w, frames, r, err
}
