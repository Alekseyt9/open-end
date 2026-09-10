package observer

import (
	"open-end/internal/world"
	"strings"
	"testing"
)

func TestSymbolMetricsOwnershipAndIntegrity(t *testing.T) {
	hash := strings.Repeat("a", 64)
	w := &world.World{Config: world.Config{Symbols: "persistent"}, Cells: make([]world.Cell, 4), Symbols: &world.SymbolState{SymbolCounts: world.SymbolCounts{Writes: 1}, Actors: map[string]*world.SymbolCounts{hash: {Writes: 1}}}}
	first := observeSymbols(w)
	w.Symbols.Writes++
	w.Symbols.Actors[hash].Writes++
	last := observeSymbols(w)
	if first.State.Actors[hash].Writes != 1 {
		t.Fatal("metrics alias world")
	}
	frames := []Metrics{{Symbols: first}, {Symbols: last}}
	r, err := summarizeSymbols(frames)
	if err != nil || r.Writes != 1 || len(r.Activity) != 1 {
		t.Fatalf("summary: %+v, %v", r, err)
	}
	last.State.Actors[hash].Writes++
	if _, err := summarizeSymbols(frames); err == nil {
		t.Fatal("bad actor ledger accepted")
	}
	last.State.Actors[hash].Writes--
	last.State.Reads = 1
	if _, err := summarizeSymbols(frames); err == nil {
		t.Fatal("bad totals accepted")
	}
	last.State.Reads = 0
	last.Words = 5
	if _, err := summarizeSymbols(frames); err == nil {
		t.Fatal("unbounded words accepted")
	}
	last.Words = 0
	last.Model = "scrambled"
	if _, err := summarizeSymbols(frames); err == nil {
		t.Fatal("mode change in session accepted")
	}
}
