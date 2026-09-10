package observer_test

import (
	"open-end/internal/observer"
	"testing"
)

func TestVariationTelemetryReconcilesAndOwnsRecords(t *testing.T) {
	w := ecology(t)
	w.Config.CopyModel = "evolving"
	frames := capture(t, w, 3000, 250)
	summary, err := observer.Summarize(frames, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Variation == nil || summary.Variation.CodeCopies != summary.Copies || summary.Copies == 0 {
		t.Fatal("variation window does not reconcile")
	}
	before := observer.Observe(w)
	for _, rec := range w.Variation {
		rec.Copies++
	}
	after := observer.Observe(w)
	if before.Variation[0].Copies+1 != after.Variation[0].Copies {
		t.Fatal("metrics alias mutable ledger")
	}
}
