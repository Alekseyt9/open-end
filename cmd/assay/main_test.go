package main

import (
	"bytes"
	"encoding/json"
	"open-end/internal/evolution"
	"open-end/internal/experiment"
	"open-end/internal/vm"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkerReplayAndDensityControls(t *testing.T) {
	dir := t.TempDir()
	a, b := vm.EcologySeed(), vm.EcologySeed()
	b[8] = vm.Instruction{Op: vm.BIND}
	pair := experiment.Pair{A: experiment.Template{Hash: evolution.Hash(a, [8]int{}), Code: a}, B: experiment.Template{Hash: evolution.Hash(b, [8]int{}), Code: b}}
	data, _ := json.Marshal(pair)
	pairPath := filepath.Join(dir, "pair.json")
	if err := os.WriteFile(pairPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	var results []Result
	for _, name := range []string{"first", "second"} {
		var output bytes.Buffer
		if err := run([]string{"-pair", pairPath, "-worker", "mixed", "-ticks", "500", "-every", "100", "-output", filepath.Join(dir, name)}, &output); err != nil {
			t.Fatal(err)
		}
		var r Result
		if err := json.Unmarshal(output.Bytes(), &r); err != nil {
			t.Fatal(err)
		}
		results = append(results, r)
	}
	if results[0].StateHash != results[1].StateHash || results[0].InitialA != 16 || results[0].InitialB != 16 {
		t.Fatal("assay is not reproducible or founder counts differ")
	}
	if results[0].CopiesA == 0 || results[0].CopiesB == 0 {
		t.Fatal("fixture lineages did not produce offspring")
	}
}

func TestInvalidAssayFailsBeforeOutput(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "output")
	if err := run([]string{"-founders", "3", "-output", dir}, &bytes.Buffer{}); err == nil {
		t.Fatal("odd inoculum accepted")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("invalid request created output")
	}
}
