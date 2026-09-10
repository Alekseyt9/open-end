package main

import (
	"bytes"
	"encoding/json"
	"open-end/internal/observer"
	"os"
	"path/filepath"
	"testing"
)

func fixtureBatch(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write := func(name string, v any) {
		t.Helper()
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("manifest.json", map[string]any{"Status": "complete", "Completed": 1, "Total": 1})
	write("summary.json", []runSummary{{Case: "test", Seed: 1, Tick: 30000, Entities: 20, Genomes: 1, RecentCopies: 10, RecentTicks: 10000, Metrics: "run.jsonl"}})
	var data bytes.Buffer
	enc := json.NewEncoder(&data)
	for i := 0; i < 4; i++ {
		if err := enc.Encode(observer.Metrics{Tick: uint64(i * 10000), Entities: 20, Genomes: 1, ActiveGenomes: []observer.Genome{{Hash: "a", Count: 20, Frequency: 1, Copies: uint64(i * 10), Instructions: uint64(i * 100), Moved: uint64(i * 10)}}}); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "run.jsonl"), data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAnalyzeCompletedBatch(t *testing.T) {
	dir := fixtureBatch(t)
	var out bytes.Buffer
	if err := run([]string{"-input", dir}, &out); err != nil {
		t.Fatal(err)
	}
	var result []runAnalysis
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || len(result[0].Persistence.Genomes) != 1 || result[0].Persistence.Genomes[0].Copies != 30 {
		t.Fatalf("unexpected analysis: %s", out.String())
	}
}

func TestAnalyzeRejectsIncompleteOrMismatchedBatch(t *testing.T) {
	dir := fixtureBatch(t)
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"Status":"running","Completed":1,"Total":2}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-input", dir}, &bytes.Buffer{}); err == nil {
		t.Fatal("partial batch accepted")
	}
	dir = fixtureBatch(t)
	if err := os.WriteFile(filepath.Join(dir, "summary.json"), []byte(`[]`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-input", dir}, &bytes.Buffer{}); err == nil {
		t.Fatal("manifest mismatch accepted")
	}
}

func TestReadFramesRejectsBrokenTimeline(t *testing.T) {
	dir := fixtureBatch(t)
	frames, err := readFrames(filepath.Join(dir, "run.jsonl"), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 2 || frames[0].Tick != 20000 {
		t.Fatal("tail was not bounded")
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.jsonl"), []byte("{\"tick\":10}\n{\"tick\":5}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readFrames(filepath.Join(dir, "broken.jsonl"), 2); err == nil {
		t.Fatal("decreasing tick accepted")
	}
}
