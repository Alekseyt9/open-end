package council

import (
	"bytes"
	"encoding/json"
	"io"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) (string, string, Request, Response) {
	t.Helper()
	root := t.TempDir()
	input := filepath.Join(root, "batch")
	if err := os.Mkdir(input, 0755); err != nil {
		t.Fatal(err)
	}
	rows := []map[string]any{}
	for seed := uint64(1); seed <= 2; seed++ {
		cfg := world.DefaultConfig()
		cfg.Width = 12
		cfg.Height = 12
		cfg.MaxEntities = 144
		cfg.Ecology = true
		cfg.MatterDiffusion = 4
		cfg.ChemicalDiffusion = 4
		cfg.Seed = seed
		w, err := world.New(cfg)
		if err != nil {
			t.Fatal(err)
		}
		tr := observer.NewTracker(w)
		var metrics, snapshot bytes.Buffer
		enc := json.NewEncoder(&metrics)
		_ = enc.Encode(tr.Frame(w))
		for i := 0; i < 200; i++ {
			kernel.StepObserved(w, tr)
			if (i+1)%50 == 0 {
				_ = enc.Encode(tr.Frame(w))
			}
		}
		if err := kernel.Save(&snapshot, w); err != nil {
			t.Fatal(err)
		}
		name := "one"
		if seed == 2 {
			name = "two"
		}
		if err := writeNew(filepath.Join(input, name+".jsonl"), metrics.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := writeNew(filepath.Join(input, name+".json"), snapshot.Bytes()); err != nil {
			t.Fatal(err)
		}
		m := observer.Observe(w)
		rows = append(rows, map[string]any{"Case": "ecology", "Seed": seed, "Tick": w.Tick, "Entities": m.Entities, "Genomes": m.Genomes, "Snapshot": name + ".json", "Metrics": name + ".jsonl", "StateSHA256": kernel.Hash(w)})
	}
	if err := writeJSON(filepath.Join(input, "manifest.json"), map[string]any{"Status": "complete", "Completed": 2, "Total": 2}); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(input, "summary.json"), rows); err != nil {
		t.Fatal(err)
	}
	round := filepath.Join(root, "round")
	r, err := Prepare(input, round, 200)
	if err != nil {
		t.Fatal(err)
	}
	a := Response{Version: 1, RequestID: r.ID, Author: "test fixture", Claims: []Claim{}, Proposals: []Proposal{}}
	for _, topic := range []string{"dominance", "niches", "structures", "stagnation"} {
		a.Claims = append(a.Claims, Claim{Topic: topic, Kind: "hypothesis", Text: "Synthetic hypothesis", Caveat: "Test only", Evidence: []string{r.Worlds[0].Facts[0].ID}})
	}
	p := Proposal{ID: "recycle-y", BaseRules: "builtin", Mechanism: "Y and field energy regenerate X", Rationale: "Test a new reaction", Evidence: []string{r.Worlds[0].Facts[0].ID}, Prediction: "Reaction flow changes", Risk: "Could collapse the population", Module: builtinSource()}
	p.Module.Version = "recycle-y-v1"
	p.Module.Rules[1].Name = "recycle-y"
	p.Module.Rules[1].Consume = map[string]int{"Y": 1, "field_energy": 4}
	p.Module.Rules[1].Produce = map[string]int{"X": 1}
	a.Proposals = append(a.Proposals, p)
	return input, round, r, a
}
func saveResponse(t *testing.T, round string, a Response) string {
	t.Helper()
	path := filepath.Join(round, "answer.json")
	b, _ := json.Marshal(a)
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPrepareCheckAndTrialReplay(t *testing.T) {
	input, round, r, a := fixture(t)
	path := saveResponse(t, round, a)
	checked, err := Check(round, path)
	if err != nil {
		t.Fatal(err)
	}
	if checked.Kinds["recycle-y"] != "structural" {
		t.Fatal("wrong classification")
	}
	var review bytes.Buffer
	if err := Review(&review, checked); err != nil || review.Len() == 0 {
		t.Fatal("missing review", err)
	}
	if _, err := Check(round, filepath.Join(round, "response.template.json")); err == nil {
		t.Fatal("blank template accepted")
	}
	if _, err := Prepare(input, round, 200); err == nil {
		t.Fatal("existing round overwritten")
	}
	before := map[string]string{}
	for _, wb := range r.Worlds {
		b, _ := os.ReadFile(filepath.Join(round, wb.ID+".snapshot.json"))
		before[wb.ID] = digest(b)
	}
	opts := TrialOptions{Ticks: 100, Every: 20, Window: 80, Workers: 1}
	first, err := Trial(round, path, filepath.Join(t.TempDir(), "trial"), opts, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	opts.Workers = 2
	dest := filepath.Join(t.TempDir(), "trial")
	second, err := Trial(round, path, dest, opts, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 4 || len(second) != 4 {
		t.Fatal("missing paired branches")
	}
	for i := range first {
		if first[i].FinalHash != second[i].FinalHash {
			t.Fatal("worker count changes physics")
		}
		if first[i].Variant == "control" {
			w, err := loadWorld(filepath.Join(round, first[i].WorldID+".snapshot.json"))
			if err != nil {
				t.Fatal(err)
			}
			for n := 0; n < 100; n++ {
				kernel.Step(w)
			}
			if kernel.Hash(w) != first[i].FinalHash {
				t.Fatal("control differs from direct kernel continuation")
			}
		}
	}
	for _, wb := range r.Worlds {
		b, _ := os.ReadFile(filepath.Join(round, wb.ID+".snapshot.json"))
		if digest(b) != before[wb.ID] {
			t.Fatal("source snapshot modified")
		}
	}
	var manifest trialManifest
	if err := readJSON(filepath.Join(dest, "manifest.json"), &manifest); err != nil || manifest.Status != "complete" || manifest.Completed != 4 {
		t.Fatal("bad completion", err)
	}
	if _, err := Trial(round, path, dest, opts, io.Discard); err == nil {
		t.Fatal("trial output overwritten")
	}
	nextRound := filepath.Join(t.TempDir(), "next")
	next, err := PrepareVariant(dest, nextRound, 80, "recycle-y")
	if err != nil || len(next.Worlds) != 2 || next.Worlds[0].Tick != 300 {
		t.Fatal("could not prepare next round", err)
	}
	if next.Worlds[0].Rules.Hash == "builtin" {
		t.Fatal("next round lost installed rules")
	}
}

func TestResponseRejections(t *testing.T) {
	_, round, _, original := fixture(t)
	cases := map[string]func(*Response){
		"stale":           func(a *Response) { a.RequestID = "old" },
		"invented fact":   func(a *Response) { a.Claims[0].Evidence = []string{"nonexistent"} },
		"missing caveat":  func(a *Response) { a.Claims[0].Caveat = "" },
		"missing topic":   func(a *Response) { a.Claims[0].Topic = "niches" },
		"path proposal":   func(a *Response) { a.Proposals[0].ID = "../escape" },
		"wrong base":      func(a *Response) { a.Proposals[0].BaseRules = "other" },
		"energy creation": func(a *Response) { a.Proposals[0].Module.Rules[1].Produce["energy"] = 100 },
		"cosmetic":        func(a *Response) { a.Proposals[0].Module = builtinSource(); a.Proposals[0].Module.Version = "renamed" },
		"unreachable": func(a *Response) {
			extra := a.Proposals[0].Module.Rules[1]
			extra.ID = 2
			a.Proposals[0].Module = builtinSource()
			a.Proposals[0].Module.Rules = append(a.Proposals[0].Module.Rules, extra)
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			b, _ := json.Marshal(original)
			var a Response
			_ = json.Unmarshal(b, &a)
			change(&a)
			if _, err := Check(round, saveResponse(t, round, a)); err == nil {
				t.Fatal("invalid response accepted")
			}
		})
	}
	path := saveResponse(t, round, original)
	b, _ := os.ReadFile(path)
	b = bytes.Replace(b, []byte(`"version":1`), []byte(`"version":1,"version":1`), 1)
	_ = os.WriteFile(path, b, 0644)
	if _, err := Check(round, path); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
	original.Proposals = nil
	if _, err := Check(round, saveResponse(t, round, original)); err != nil {
		t.Fatal("observer-only response rejected", err)
	}
}

func TestFrozenArtifactsAndFactProvenance(t *testing.T) {
	_, round, r, a := fixture(t)
	path := saveResponse(t, round, a)
	evidencePath := filepath.Join(round, r.Worlds[0].ID+".evidence.json")
	b, _ := os.ReadFile(evidencePath)
	_ = os.WriteFile(evidencePath, append(b, ' '), 0644)
	if _, err := Check(round, path); err == nil {
		t.Fatal("modified evidence accepted")
	}
	_ = os.WriteFile(evidencePath, b, 0644)
	snapshotPath := filepath.Join(round, r.Worlds[0].ID+".snapshot.json")
	originalSnapshot, _ := os.ReadFile(snapshotPath)
	w, err := loadWorld(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	w.Config.MutationPPM++
	var altered bytes.Buffer
	if err := kernel.Save(&altered, w); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(snapshotPath, altered.Bytes(), 0644)
	if _, err := Check(round, path); err == nil {
		t.Fatal("modified frozen snapshot accepted")
	}
	_ = os.WriteFile(snapshotPath, originalSnapshot, 0644)
	r.Worlds[0].Facts[0].Value = json.RawMessage(`999`)
	r.ID = ""
	r.ID = jsonHash(r)
	a.RequestID = r.ID
	encoded, _ := json.Marshal(r)
	_ = os.WriteFile(filepath.Join(round, "request.json"), encoded, 0644)
	if _, err := Check(round, saveResponse(t, round, a)); err == nil {
		t.Fatal("false fact accepted with recomputed request hash")
	}
}

func TestPrepareRejectsPartialAndMismatchedBatch(t *testing.T) {
	input, _, _, _ := fixture(t)
	_ = os.WriteFile(filepath.Join(input, "manifest.json"), []byte(`{"Status":"running","Completed":2,"Total":2}`), 0644)
	if _, err := Prepare(input, filepath.Join(t.TempDir(), "round"), 200); err == nil {
		t.Fatal("partial batch accepted")
	}
	_ = os.WriteFile(filepath.Join(input, "manifest.json"), []byte(`{"Status":"complete","Completed":2,"Total":2}`), 0644)
	b, _ := os.ReadFile(filepath.Join(input, "summary.json"))
	b = bytes.Replace(b, []byte(`"Seed": 1`), []byte(`"Seed": 999`), 1)
	_ = os.WriteFile(filepath.Join(input, "summary.json"), b, 0644)
	if _, err := Prepare(input, filepath.Join(t.TempDir(), "round"), 200); err == nil {
		t.Fatal("mismatched batch accepted")
	}
}
