package observer

import (
	"open-end/internal/world"
	"strings"
	"testing"
)

func TestEnvironmentSummaryIntegrityAndOwnership(t *testing.T) {
	hash := strings.Repeat("a", 64)
	w := &world.World{Config: world.Config{Environment: "coupled", Width: 2, Height: 2}, Cells: make([]world.Cell, 4), Environment: &world.EnvironmentState{Built: 2, Emitted: 10, Actors: map[string]*world.EnvironmentActor{hash: {Built: 2, Emitted: 10}}}}
	w.Cells[0].Terrain = 2
	w.Cells[0].Signal = 10
	first := observeEnvironment(w)
	w.Environment.Built++
	w.Environment.Actors[hash].Built++
	w.Cells[0].Terrain++
	w.Environment.Decayed++
	w.Cells[0].Signal--
	last := observeEnvironment(w)
	if first.State.Actors[hash].Built != 2 || first.Fields[0].Terrain != 2 {
		t.Fatal("aliased telemetry")
	}
	frames := []Metrics{{Environment: first}, {Environment: last}}
	r, err := summarizeEnvironment(frames)
	if err != nil {
		t.Fatal(err)
	}
	if r.Built != 1 || r.Decayed != 1 || r.Terrain != 3 || r.Signal != 9 || len(r.Activity) != 1 {
		t.Fatal("incorrect environment delta")
	}
	r.Fields[0].Terrain = 0
	if last.Fields[0].Terrain != 3 {
		t.Fatal("summary aliases metrics")
	}
	last.State.Actors[hash].Built = 2
	if _, err := summarizeEnvironment(frames); err == nil {
		t.Fatal("unreconciled actor accepted")
	}
	last.State.Actors[hash].Built = 3
	last.Fields = append(last.Fields, last.Fields[0])
	if _, err := summarizeEnvironment(frames); err == nil {
		t.Fatal("duplicate field accepted")
	}
}
