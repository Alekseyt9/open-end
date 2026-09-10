package experiment

import (
	"bytes"
	"fmt"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
)

// SymbolTrial uses the same provenance/telemetry envelope as environment trials.
type SymbolTrial = EnvironmentTrial

// ContinueSymbols intervenes only on reception. Write costs, words, memories,
// instruction repertoire and both RNGs are identical at the branch point.
func ContinueSymbols(source *world.World, mode string, ticks, every int) (*world.World, []observer.Metrics, SymbolTrial, error) {
	r := SymbolTrial{Variant: mode}
	if source == nil || source.Config.Symbols != "persistent" || ticks < 1 || every < 1 || (mode != "persistent" && mode != "scrambled" && mode != "unreadable") {
		return nil, nil, r, fmt.Errorf("invalid symbol continuation: requires persistent source")
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
	w.Config.Symbols = mode
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
