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

func TestCLICopyModelReplay(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "encoded.json")
	var first, resumed, full bytes.Buffer
	base := []string{"-width", "12", "-height", "12", "-copy-model", "evolving"}
	if err := run(append(append([]string{}, base...), "-ticks", "200", "-save", path), &first); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-load", path, "-ticks", "100"}, &resumed); err != nil {
		t.Fatal(err)
	}
	if err := run(append(append([]string{}, base...), "-ticks", "300"), &full); err != nil {
		t.Fatal(err)
	}
	hash := func(b *bytes.Buffer) string { s := strings.Split(b.String(), "state_sha256="); return s[len(s)-1] }
	if hash(&resumed) != hash(&full) {
		t.Fatal("encoded CLI replay differs")
	}
	if err := run([]string{"-load", path, "-copy-model", "fixed"}, &first); err == nil {
		t.Fatal("copy model override accepted")
	}
	if err := run([]string{"-copy-model", "typo", "-ticks", "0"}, &first); err == nil {
		t.Fatal("unknown model accepted")
	}
}

func TestCLIEnvironmentReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "environment.json")
	var first, resumed, full bytes.Buffer
	base := []string{"-width", "8", "-height", "8", "-ecology", "-environment", "coupled", "-copy-model", "evolving"}
	if err := run(append(append([]string{}, base...), "-ticks", "100", "-save", path), &first); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"-load", path, "-ticks", "100"}, &resumed); err != nil {
		t.Fatal(err)
	}
	if err := run(append(append([]string{}, base...), "-ticks", "200"), &full); err != nil {
		t.Fatal(err)
	}
	hash := func(b *bytes.Buffer) string { s := strings.Split(b.String(), "state_sha256="); return s[len(s)-1] }
	if hash(&resumed) != hash(&full) {
		t.Fatal("environment CLI replay differs")
	}
	if err := run([]string{"-load", path, "-environment", "inert"}, &first); err == nil {
		t.Fatal("environment override accepted")
	}
	if err := run([]string{"-environment", "coupled"}, &first); err == nil {
		t.Fatal("engineering without ecology accepted")
	}
}

func TestCLIRulesReplayWithoutSourceFiles(t *testing.T) {
	dir := t.TempDir()
	base, replacement := filepath.Join(dir, "base.json"), filepath.Join(dir, "replacement.json")
	for path, name := range map[string]string{base: "baseline", replacement: "direct-x"} {
		b, err := os.ReadFile("../../examples/rules/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	common := []string{"-width", "12", "-height", "12", "-ecology", "-rules", base, "-rule-change", "100=" + replacement, "-rollback-at", "200"}
	var full, first, resumed bytes.Buffer
	if err := run(append(append([]string{}, common...), "-ticks", "300"), &full); err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(dir, "state.json")
	if err := run(append(append([]string{}, common...), "-ticks", "50", "-save", saved), &first); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{base, replacement} {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := run([]string{"-load", saved, "-ticks", "250"}, &resumed); err != nil {
		t.Fatal(err)
	}
	hash := func(b *bytes.Buffer) string { s := strings.Split(b.String(), "state_sha256="); return s[len(s)-1] }
	if hash(&full) != hash(&resumed) {
		t.Fatal("frozen rule replay differs")
	}
	if err := run([]string{"-load", saved, "-rollback-at", "500"}, &first); err == nil {
		t.Fatal("accepted override of snapshot rules")
	}
}

func TestCLIBadRulesFailBeforeCreatingMetrics(t *testing.T) {
	for _, extra := range [][]string{
		{"-rules", "../../examples/rules/baseline.json"},
		{"-ecology", "-rollback-at", "10"},
		{"-ecology", "-rule-change", "bad"},
		{"-ecology", "-rule-change", "10=../../examples/rules/baseline.json", "-rollback-at", "10"},
	} {
		path := filepath.Join(t.TempDir(), "metrics.jsonl")
		var out bytes.Buffer
		if err := run(append([]string{"-ticks", "0", "-metrics", path}, extra...), &out); err == nil {
			t.Fatal("accepted invalid rule flags")
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("created metrics before rejecting rules")
		}
	}
}
