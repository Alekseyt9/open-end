package experiment

import (
	"bytes"
	"fmt"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
)

// ContinueCollectives applies a single recorded ablation and observes groups.
// Reference groups always come from the unchanged source, even after unbinding.
func ContinueCollectives(source *world.World, mode string, ticks, every int, minAge uint64) (*world.World, []observer.Metrics, EnvironmentTrial, error) {
	r := EnvironmentTrial{Variant: mode}
	if ticks < 1 || every < 1 || minAge == 0 {
		return nil, nil, r, fmt.Errorf("invalid collective continuation")
	}
	w, err := CollectiveStart(source, mode)
	if err != nil {
		return nil, nil, r, err
	}
	r.SourceHash = kernel.Hash(source)
	r.Seed = source.Config.Seed
	r.InitialHash = kernel.Hash(w)
	tracker := observer.NewTracker(w)
	if err := tracker.EnableCollectives(w, source, minAge); err != nil {
		return nil, nil, r, err
	}
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

// CollectiveStart clones a validated source and applies one explicit treatment.
// Analysis replays use exactly the same initialization as the original assay.
func CollectiveStart(source *world.World, mode string) (*world.World, error) {
	if source == nil || source.Config.CollectiveAblation != "" || (mode != "intact" && mode != "bonds" && mode != "sharing" && mode != "signal-reading") {
		return nil, fmt.Errorf("invalid collective source or treatment")
	}
	var buf bytes.Buffer
	if err := kernel.Save(&buf, source); err != nil {
		return nil, err
	}
	w, err := kernel.Load(&buf)
	if err != nil {
		return nil, err
	}
	if mode != "intact" {
		w.Config.CollectiveAblation = mode
	}
	if mode == "bonds" {
		w.Relations = map[string]world.Relation{}
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	return w, nil
}
