package council

import (
	"bytes"
	"encoding/json"
	"fmt"
	"open-end/internal/dsl"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

type Fact struct {
	ID      string          `json:"id"`
	Pointer string          `json:"evidence_pointer"`
	Value   json.RawMessage `json:"value"`
}
type WorldBrief struct {
	BondMotion          string                `json:"bond_motion,omitempty"`
	MetabolicSwitchCost int                   `json:"metabolic_switch_cost,omitempty"`
	Symbols             string                `json:"symbols,omitempty"`
	CollectiveAblation  string                `json:"collective_ablation,omitempty"`
	CollectiveAge       uint64                `json:"collective_age,omitempty"`
	Environment         string                `json:"environment,omitempty"`
	CopyModel           string                `json:"copy_model,omitempty"`
	ID                  string                `json:"id"`
	Case                string                `json:"case"`
	Seed                uint64                `json:"seed"`
	Tick                uint64                `json:"tick"`
	MutationPPM         int                   `json:"mutation_ppm"`
	SnapshotHash        string                `json:"snapshot_sha256"`
	MetricsHash         string                `json:"metrics_file_sha256"`
	EvidenceHash        string                `json:"evidence_file_sha256"`
	Rules               observer.RuleIdentity `json:"active_rules"`
	RuleSource          dsl.Document          `json:"active_rule_source"`
	Facts               []Fact                `json:"facts"`
}
type Request struct {
	Version     int          `json:"version"`
	ID          string       `json:"request_sha256"`
	Kernel      string       `json:"kernel_version"`
	RuleVersion string       `json:"rule_version"`
	Limits      []string     `json:"capabilities_and_limits"`
	Worlds      []WorldBrief `json:"worlds"`
}
type Evidence struct {
	Summary  observer.WindowSummary `json:"summary"`
	Dynamics observer.Dynamics      `json:"dynamics"`
}

func builtinSource() dsl.Document {
	return dsl.Document{Format: 1, Version: "builtin-equivalent", InstructionBudget: 32, Rules: []dsl.RuleSpec{
		{ID: 0, Name: "x-to-y", Consume: map[string]int{"X": 1}, Produce: map[string]int{"Y": 1, "energy": 4}, EnergyCost: 1, MaxBatch: 64},
		{ID: 1, Name: "y-to-z", Consume: map[string]int{"Y": 1}, Produce: map[string]int{"Z": 1, "energy": 4}, EnergyCost: 1, MaxBatch: 64},
	}}
}
func loadWorld(path string) (*world.World, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return kernel.Load(f)
}

// Prepare freezes a completed experiment into a new round directory.
func Prepare(input, out string, window uint64) (Request, error) {
	return PrepareVariant(input, out, window, "")
}

// PrepareVariant also accepts one selected variant of a completed trial, so
// the next discussion round uses the same protocol as the initial experiment.
func PrepareVariant(input, out string, window uint64, variant string) (Request, error) {
	r := Request{Version: 1, Kernel: kernel.Version, RuleVersion: kernel.RuleVersion, Worlds: []WorldBrief{}, Limits: []string{
		"Ответ готовит человек или AI в чате; сетевого API и автоматического вызова модели нет.",
		"Наблюдения должны ссылаться на ID фактов. Ниши, кооперация, причинность и адаптивная ценность требуют гипотез и дополнительных проверок.",
		"DSL: только локальные реакции X/Y/Z/field_energy/energy; масса и энергия сохраняются; цена инструкции 1..64, до 16 правил и 64 единиц бюджета.",
		"Мутации CONVERT создают только ID 0 и 1. Добавленный ID без изменения доступного механизма может остаться неиспользуемым.",
		"Предлагать общий физический механизм, не редактировать геномы, особей, ресурсы или RNG. Проверка предложения не устанавливает его в исходные миры.",
		"Поведенческие хеши агрегированы по миру; novelty detector — эвристика, не доказательство адаптивной новизны.",
	}}
	var manifest struct {
		Status                                   string
		Completed, Total                         int
		Workers, GOMAXPROCS                      int
		Seeds                                    []int
		Cases                                    []string
		Ticks, Every                             int
		StartedUTC, SourceRevision, BinarySHA256 string
		SourceDirty                              bool
		ElapsedSeconds                           float64
	}
	// Runner metadata may acquire new fields; only the completion gate is read.
	b, err := os.ReadFile(filepath.Join(input, "manifest.json"))
	if err != nil {
		return r, err
	}
	if err = json.Unmarshal(b, &manifest); err != nil {
		return r, err
	}
	if manifest.Status != "complete" || manifest.Total < 1 || manifest.Total != manifest.Completed {
		return r, fmt.Errorf("experiment is not complete")
	}
	type sourceRow struct {
		Case                           string
		Seed, Tick                     uint64
		Entities, Genomes              int
		Snapshot, Metrics, StateSHA256 string
	}
	rows := []sourceRow{}
	if variant != "" {
		if !identifier.MatchString(variant) {
			return r, fmt.Errorf("invalid variant")
		}
		var trialRows []TrialRow
		if err := readJSON(filepath.Join(input, "results.json"), &trialRows); err != nil {
			return r, err
		}
		if len(trialRows) != manifest.Total {
			return r, fmt.Errorf("manifest/results mismatch")
		}
		for _, row := range trialRows {
			if row.Error != "" {
				return r, fmt.Errorf("trial contains failed branches")
			}
			if row.Variant == variant {
				rows = append(rows, sourceRow{Case: row.Case, Seed: row.Seed, Tick: row.Summary.ToTick, Entities: int(row.Summary.Population.End), Genomes: int(row.Summary.Genomes.End), Snapshot: row.Snapshot, Metrics: row.Metrics, StateSHA256: row.FinalHash})
			}
		}
		if len(rows) == 0 {
			return r, fmt.Errorf("variant has no branches")
		}
	} else {
		b, err = os.ReadFile(filepath.Join(input, "summary.json"))
		if err != nil {
			return r, fmt.Errorf("read batch summary (for council trial input, select -variant): %w", err)
		}
		if err = json.Unmarshal(b, &rows); err != nil {
			return r, err
		}
		if len(rows) != manifest.Total {
			return r, fmt.Errorf("manifest/summary mismatch")
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Case != rows[j].Case {
			return rows[i].Case < rows[j].Case
		}
		return rows[i].Seed < rows[j].Seed
	})
	type artifact struct {
		name string
		data []byte
	}
	files := []artifact{}
	seen := map[string]bool{}
	for i, row := range rows {
		if !leaf(row.Metrics) || !leaf(row.Snapshot) || seen[row.Metrics] || seen[row.Snapshot] {
			return r, fmt.Errorf("invalid/duplicate source filename")
		}
		seen[row.Metrics] = true
		seen[row.Snapshot] = true
		w, err := loadWorld(filepath.Join(input, row.Snapshot))
		if err != nil {
			return r, err
		}
		if !w.Config.Ecology {
			return r, fmt.Errorf("%s: ecology is required for this DSL interface", row.Case)
		}
		if w.RuleState != nil && len(w.RuleState.Pending) > 0 {
			return r, fmt.Errorf("pending rule transitions must finish before preparing a round")
		}
		if kernel.Hash(w) != row.StateSHA256 || w.Config.Seed != row.Seed || w.Tick != row.Tick {
			return r, fmt.Errorf("snapshot and batch summary disagree")
		}
		data, err := os.ReadFile(filepath.Join(input, row.Metrics))
		if err != nil {
			return r, err
		}
		frames, err := observer.ReadWindow(bytes.NewReader(data), window)
		if err != nil {
			return r, err
		}
		s, err := observer.Summarize(frames, window)
		if err != nil {
			return r, err
		}
		last := frames[len(frames)-1]
		last.Telemetry = nil
		last.Interval = observer.Interval{}
		if !reflect.DeepEqual(last, observer.Observe(w)) || s.Seed != row.Seed || s.Population.End != int64(row.Entities) || s.Genomes.End != int64(row.Genomes) {
			return r, fmt.Errorf("telemetry and snapshot disagree")
		}
		d, err := observer.DetectDynamics(frames, window, observer.DefaultDynamicsConfig())
		if err != nil {
			return r, err
		}
		id := fmt.Sprintf("w%03d", i+1)
		e := Evidence{s, d}
		encoded, _ := json.MarshalIndent(e, "", "  ")
		encoded = append(encoded, '\n')
		wb := WorldBrief{ID: id, Case: row.Case, Seed: row.Seed, Tick: w.Tick, MutationPPM: w.Config.MutationPPM, SnapshotHash: kernel.Hash(w), MetricsHash: digest(data), EvidenceHash: digest(encoded), Rules: s.RulesEnd, RuleSource: builtinSource(), Facts: []Fact{}}
		wb.CopyModel = w.Config.CopyModel
		wb.Environment = w.Config.Environment
		wb.Symbols = w.Config.Symbols
		wb.BondMotion = w.Config.BondMotion
		wb.MetabolicSwitchCost = w.Config.MetabolicSwitchCost
		wb.CollectiveAblation = w.Config.CollectiveAblation
		if s.Collectives != nil {
			wb.CollectiveAge = s.Collectives.End.MinAge
		}
		if w.RuleState != nil && w.RuleState.Active != nil {
			wb.RuleSource = w.RuleState.Active.Source
		}
		paths := []string{"summary/population", "summary/diversity_end", "summary/dominant_genomes_at_end", "summary/structures_end/largest", "summary/structures_end/linked_components", "summary/structures_end/mixed_genome_components", "summary/resource_flows", "summary/failed_attempts", "summary/copies", "summary/deaths", "dynamics/status", "dynamics/new_genomes_not_used_as_novelty"}
		if s.Variation != nil {
			for _, field := range []string{"model", "code_policies", "memory_policies", "persistent_code_policies", "changed_code", "recombined"} {
				paths = append(paths, "summary/variation/"+field)
			}
		}
		if s.Environment != nil {
			for _, field := range []string{"built", "emitted", "blocked_light", "attenuated_transfer", "terrain_at_end", "signal_energy_at_end"} {
				paths = append(paths, "summary/environment/"+field)
			}
		}
		if s.Collectives != nil {
			for _, field := range []string{"mean_linked_groups", "bonded_transfer_energy", "new_daughter_candidates", "new_productive_daughter_candidates"} {
				paths = append(paths, "summary/collectives/"+field)
			}
		}
		if s.Symbols != nil {
			for _, field := range []string{"writes", "nonempty_reads", "pair_reads", "foreign_reads", "context_lookups"} {
				paths = append(paths, "summary/symbols/"+field)
			}
		}
		for _, path := range paths {
			parts := strings.Split(path, "/")
			value, err := pointerValue(encoded, "/"+path)
			if err != nil {
				return r, err
			}
			wb.Facts = append(wb.Facts, Fact{id + "." + parts[len(parts)-1], "/" + path, value})
		}
		r.Worlds = append(r.Worlds, wb)
		files = append(files, artifact{id + ".evidence.json", encoded})
		var snapshot bytes.Buffer
		if err := kernel.Save(&snapshot, w); err != nil {
			return r, err
		}
		files = append(files, artifact{id + ".snapshot.json", snapshot.Bytes()})
	}
	r.ID = jsonHash(r)
	if err := os.Mkdir(out, 0755); err != nil {
		return r, err
	}
	for _, f := range files {
		if err := writeNew(filepath.Join(out, f.name), f.data); err != nil {
			return r, err
		}
	}
	if err := writeJSON(filepath.Join(out, "request.json"), r); err != nil {
		return r, err
	}
	template := Response{Version: 1, RequestID: r.ID, Author: "", Claims: []Claim{}, Proposals: []Proposal{}}
	for _, topic := range []string{"dominance", "niches", "structures", "stagnation"} {
		template.Claims = append(template.Claims, Claim{Topic: topic, Kind: "hypothesis", Evidence: []string{}, Text: "", Caveat: ""})
	}
	if err := writeJSON(filepath.Join(out, "response.template.json"), template); err != nil {
		return r, err
	}
	var brief strings.Builder
	fmt.Fprintf(&brief, "# Досье AI-наблюдателя\n\nЗапрос `%s`. Миры: %d.\n\n", r.ID, len(r.Worlds))
	for _, line := range r.Limits {
		fmt.Fprintf(&brief, "- %s\n", line)
	}
	brief.WriteString("\nЗаполни `response.json` по `response.template.json`: автор, четыре темы, наблюдения/гипотезы и ссылки на ID фактов из `request.json`. Незаполненный шаблон не проходит проверку. Для каждой гипотезы укажи ограничение или проверку в `caveat`. Полные данные доступны в `wNNN.evidence.json`.\n\nПредложения необязательны. Формат одного предложения: `id`, `base_rules_sha256`, `mechanism`, `rationale`, `evidence` (ID фактов), `prediction`, `risk`, `module` (полный DSL document). Допустимо до четырёх предложений. Для встроенных правил base hash равен `builtin`. Никакие команды или пути из ответа не исполняются.\n\n")
	for _, w := range r.Worlds {
		fmt.Fprintf(&brief, "## %s — %s, seed %d\n\n", w.ID, w.Case, w.Seed)
		for _, f := range w.Facts {
			fmt.Fprintf(&brief, "- `%s`: %s\n", f.ID, compact(f.Value))
		}
	}
	return r, writeNew(filepath.Join(out, "brief.md"), []byte(brief.String()))
}
func compact(b []byte) string { var out bytes.Buffer; _ = json.Compact(&out, b); return out.String() }
