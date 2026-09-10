package council

import (
	"encoding/json"
	"fmt"
	"io"
	"open-end/internal/dsl"
	"open-end/internal/kernel"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Claim struct {
	Topic    string   `json:"topic"`
	Kind     string   `json:"kind"`
	Text     string   `json:"text"`
	Evidence []string `json:"evidence"`
	Caveat   string   `json:"caveat"`
}
type Proposal struct {
	ID         string       `json:"id"`
	BaseRules  string       `json:"base_rules_sha256"`
	Mechanism  string       `json:"mechanism"`
	Rationale  string       `json:"rationale"`
	Evidence   []string     `json:"evidence"`
	Prediction string       `json:"prediction"`
	Risk       string       `json:"risk"`
	Module     dsl.Document `json:"module"`
}
type Response struct {
	Version   int        `json:"version"`
	RequestID string     `json:"request_sha256"`
	Author    string     `json:"author"`
	Claims    []Claim    `json:"claims"`
	Proposals []Proposal `json:"proposals"`
}
type Checked struct {
	Request      Request
	Response     Response
	ResponseHash string
	Modules      map[string]*dsl.Module
	Kinds        map[string]string
}

func pointerValue(b []byte, path string) (json.RawMessage, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("invalid fact pointer")
	}
	var raw json.RawMessage = b
	for _, key := range strings.Split(path[1:], "/") {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return nil, err
		}
		v, ok := fields[key]
		if !ok {
			return nil, fmt.Errorf("missing evidence field")
		}
		raw = v
	}
	return raw, nil
}

func Check(round, responsePath string) (Checked, error) {
	return checkRound(round, responsePath, false)
}

func checkRound(round, responsePath string, continuation bool) (Checked, error) {
	c := Checked{Modules: map[string]*dsl.Module{}, Kinds: map[string]string{}}
	if err := readJSON(filepath.Join(round, "request.json"), &c.Request); err != nil {
		return c, err
	}
	r := c.Request
	id := r.ID
	r.ID = ""
	if r.Version != 1 || jsonHash(r) != id || r.Kernel != kernel.Version || r.RuleVersion != kernel.RuleVersion || len(r.Worlds) == 0 {
		return c, fmt.Errorf("invalid or stale request identity")
	}
	facts := map[string]bool{}
	worldIDs := map[string]bool{}
	for _, wb := range r.Worlds {
		if !identifier.MatchString(wb.ID) || worldIDs[wb.ID] {
			return c, fmt.Errorf("invalid/duplicate world ID")
		}
		worldIDs[wb.ID] = true
		w, err := loadWorld(filepath.Join(round, wb.ID+".snapshot.json"))
		if err != nil {
			return c, err
		}
		if kernel.Hash(w) != wb.SnapshotHash || w.Tick != wb.Tick || w.Config.Seed != wb.Seed || w.Config.MutationPPM != wb.MutationPPM {
			return c, fmt.Errorf("frozen snapshot changed: %s", wb.ID)
		}
		source := builtinSource()
		hash := "builtin"
		if w.RuleState != nil && w.RuleState.Active != nil {
			source = w.RuleState.Active.Source
			hash = w.RuleState.Active.Hash
		}
		if hash != wb.Rules.Hash || !reflect.DeepEqual(source, wb.RuleSource) {
			return c, fmt.Errorf("snapshot rule provenance mismatch")
		}
		data, err := os.ReadFile(filepath.Join(round, wb.ID+".evidence.json"))
		if err != nil {
			return c, err
		}
		if digest(data) != wb.EvidenceHash {
			return c, fmt.Errorf("frozen evidence changed: %s", wb.ID)
		}
		for _, fact := range wb.Facts {
			v, err := pointerValue(data, fact.Pointer)
			if err != nil {
				return c, err
			}
			if facts[fact.ID] || !strings.HasPrefix(fact.ID, wb.ID+".") || compact(v) != compact(fact.Value) {
				return c, fmt.Errorf("invalid fact provenance: %s", fact.ID)
			}
			facts[fact.ID] = true
		}
	}
	if continuation && responsePath == "" {
		return c, nil
	}
	if err := readJSON(responsePath, &c.Response); err != nil {
		return c, err
	}
	a := c.Response
	if a.Version != 1 || a.RequestID != id || strings.TrimSpace(a.Author) == "" || len(a.Claims) < 4 || len(a.Claims) > 64 || len(a.Proposals) > 4 {
		return c, fmt.Errorf("invalid response identity/author/counts")
	}
	refs := func(ids []string) error {
		if len(ids) == 0 {
			return fmt.Errorf("evidence references required")
		}
		seen := map[string]bool{}
		for _, id := range ids {
			if !facts[id] || seen[id] {
				return fmt.Errorf("unknown/duplicate evidence ID %q", id)
			}
			seen[id] = true
		}
		return nil
	}
	topics := map[string]bool{"dominance": false, "niches": false, "structures": false, "stagnation": false}
	for _, claim := range a.Claims {
		if _, ok := topics[claim.Topic]; !ok {
			return c, fmt.Errorf("unknown claim topic")
		}
		topics[claim.Topic] = true
		if (claim.Kind != "observation" && claim.Kind != "hypothesis") || strings.TrimSpace(claim.Text) == "" || len(claim.Text) > 8000 {
			return c, fmt.Errorf("invalid claim")
		}
		if claim.Kind == "hypothesis" && strings.TrimSpace(claim.Caveat) == "" {
			return c, fmt.Errorf("hypotheses require caveats or a proposed test")
		}
		if err := refs(claim.Evidence); err != nil {
			return c, err
		}
	}
	for topic, present := range topics {
		if !present {
			return c, fmt.Errorf("missing topic %s", topic)
		}
	}
	for _, p := range a.Proposals {
		if !identifier.MatchString(p.ID) || p.ID == "control" || c.Modules[p.ID] != nil {
			return c, fmt.Errorf("invalid/duplicate proposal ID")
		}
		for _, s := range []string{p.Mechanism, p.Rationale, p.Prediction, p.Risk} {
			if strings.TrimSpace(s) == "" || len(s) > 8000 {
				return c, fmt.Errorf("proposal mechanism, rationale, prediction and risk required")
			}
		}
		if err := refs(p.Evidence); err != nil {
			return c, err
		}
		module, err := dsl.Compile(p.Module)
		if err != nil {
			return c, err
		}
		kind := ""
		for _, wb := range r.Worlds {
			if p.BaseRules != wb.Rules.Hash {
				return c, fmt.Errorf("proposal %s does not match every world's base rules; prepare separate rounds", p.ID)
			}
			k, err := changeKind(wb.RuleSource, module.Source)
			if err != nil {
				return c, fmt.Errorf("%s: %w", p.ID, err)
			}
			kind = k
		}
		c.Modules[p.ID] = module
		c.Kinds[p.ID] = kind
	}
	c.ResponseHash = jsonHash(a)
	return c, nil
}

func changeKind(before, after dsl.Document) (string, error) {
	old := map[int]dsl.RuleSpec{}
	for _, r := range before.Rules {
		old[r.ID] = r
	}
	kind := "parameter"
	changed := before.InstructionBudget != after.InstructionBudget
	reachable := false
	for _, r := range after.Rules {
		b, exists := old[r.ID]
		delete(old, r.ID)
		structural := !exists || !reflect.DeepEqual(b.Consume, r.Consume) || !reflect.DeepEqual(b.Produce, r.Produce)
		different := structural || b.EnergyCost != r.EnergyCost || b.MaxBatch != r.MaxBatch
		if structural {
			kind = "structural"
		}
		changed = changed || different
		if different && (r.ID == 0 || r.ID == 1) {
			reachable = true
		}
	}
	for id := range old {
		kind = "structural"
		changed = true
		if id == 0 || id == 1 {
			reachable = true
		}
	}
	if !changed {
		return "", fmt.Errorf("proposal only renames or repeats existing rules")
	}
	if kind == "structural" && !reachable {
		return "", fmt.Errorf("new mechanism is not reachable through mutation IDs 0/1")
	}
	return kind, nil
}

func Review(out io.Writer, c Checked) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Ответ AI-наблюдателя\n\nАвтор: %s. Запрос: `%s`. Ответ: `%s`.\n\nПроверены происхождение фактов, ссылки и ограничения DSL. Смысл текста и причинные выводы требуют отдельной оценки.\n\n", c.Response.Author, c.Request.ID, c.ResponseHash)
	facts := map[string]Fact{}
	for _, w := range c.Request.Worlds {
		for _, f := range w.Facts {
			facts[f.ID] = f
		}
	}
	for _, claim := range c.Response.Claims {
		fmt.Fprintf(&b, "## %s — %s\n\n%s\n\n", claim.Topic, claim.Kind, claim.Text)
		if claim.Caveat != "" {
			fmt.Fprintf(&b, "Ограничение / проверка: %s\n\n", claim.Caveat)
		}
		for _, id := range claim.Evidence {
			fmt.Fprintf(&b, "- `%s`: %s\n", id, compact(facts[id].Value))
		}
		b.WriteString("\n")
	}
	for _, p := range c.Response.Proposals {
		fmt.Fprintf(&b, "## Предложение %s (%s)\n\n%s\n\nОснование: %s\n\nПрогноз: %s\n\nРиск: %s\n\nDSL SHA-256: `%s`.\n\n", p.ID, c.Kinds[p.ID], p.Mechanism, p.Rationale, p.Prediction, p.Risk, c.Modules[p.ID].Hash)
	}
	_, err := io.WriteString(out, b.String())
	return err
}
