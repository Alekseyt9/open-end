package main

import (
	"bytes"
	"encoding/json"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func recording(t *testing.T) []byte {
	t.Helper()
	c := world.DefaultConfig()
	c.Width = 12
	c.Height = 12
	c.MaxEntities = 144
	c.Ecology = true
	c.MatterDiffusion = 4
	c.ChemicalDiffusion = 4
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	tr := observer.NewTracker(w)
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	if err := e.Encode(tr.Frame(w)); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 300; i++ {
		kernel.StepObserved(w, tr)
		if (i+1)%100 == 0 {
			if err := e.Encode(tr.Frame(w)); err != nil {
				t.Fatal(err)
			}
		}
	}
	return b.Bytes()
}
func TestReadWindowKeepsPrecedingBoundary(t *testing.T) {
	frames, err := readWindow(bytes.NewReader(recording(t)), 150)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 3 || frames[0].Tick != 100 || frames[2].Tick != 300 {
		t.Fatal("incorrect bounded window")
	}
	r, err := observer.Summarize(frames, 150)
	if err != nil {
		t.Fatal(err)
	}
	if r.FromTick != 100 {
		t.Fatal("invented exact boundary")
	}
	if _, err := readWindow(strings.NewReader("{\"tick\":2}\n{\"tick\":1}"), 150); err == nil {
		t.Fatal("accepted reversed history")
	}
}
func TestSummaryCLIAndPartialBatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "run.jsonl")
	if err := os.WriteFile(path, recording(t), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run([]string{"-input", path, "-window", "150", "-format", "json"}, &out); err != nil {
		t.Fatal(err)
	}
	var rows []result
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Summary.FromTick != 100 {
		t.Fatal("invalid CLI result")
	}
	out.Reset()
	if err := run([]string{"-input", path, "-window", "150"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Тики 100–300") || !strings.Contains(out.String(), "Энергия:") {
		t.Fatal("missing explanation")
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"Status":"running","Completed":1,"Total":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-input", dir}, &out); err == nil {
		t.Fatal("accepted incomplete batch")
	}
}
