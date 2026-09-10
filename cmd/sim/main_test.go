package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIReplay(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	metrics := filepath.Join(dir, "metrics.jsonl")
	var first, resumed, full bytes.Buffer
	if err := run([]string{"-width", "12", "-height", "12", "-ticks", "100", "-save", path, "-metrics", metrics}, &first); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-load", path, "-ticks", "200"}, &resumed); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-width", "12", "-height", "12", "-ticks", "300"}, &full); err != nil {
		t.Fatal(err)
	}
	hash := func(b *bytes.Buffer) string { s := strings.Split(b.String(), "state_sha256="); return s[len(s)-1] }
	if hash(&resumed) != hash(&full) {
		t.Fatal("CLI replay differs")
	}
	data, err := os.ReadFile(metrics)
	if err != nil || len(bytes.Split(bytes.TrimSpace(data), []byte("\n"))) != 2 {
		t.Fatal("missing metrics")
	}
	if err := run([]string{"-load", path, "-seed", "2"}, &first); err == nil {
		t.Fatal("silently ignored seed override")
	}
	if err := run([]string{"-ticks", "-1"}, &first); err == nil {
		t.Fatal("accepted negative ticks")
	}
}
