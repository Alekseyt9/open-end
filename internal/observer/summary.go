package observer

import (
	"fmt"
	"open-end/internal/dsl"
	"sort"
)

type Range struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
	Min   int64 `json:"min_observed"`
	Max   int64 `json:"max_observed"`
}
type GenomeActivity struct {
	Hash     string    `json:"hash"`
	Start    int       `json:"start_count"`
	End      int       `json:"end_count"`
	Births   uint64    `json:"received_code"`
	Copies   uint64    `json:"produced_copies"`
	Deaths   uint64    `json:"deaths"`
	Behavior *Behavior `json:"behavior_when_baseline_known,omitempty"`
}
type Behavior struct {
	Instructions uint64   `json:"instructions"`
	Moves        uint64   `json:"moves"`
	Binds        uint64   `json:"binds"`
	Converted    [2]int64 `json:"converted_by_id"`
}
type WindowSummary struct {
	Environment     *EnvironmentWindow `json:"environment,omitempty"`
	Variation       *VariationWindow   `json:"variation,omitempty"`
	Seed            uint64             `json:"seed"`
	SessionHash     string             `json:"session_initial_world_sha256"`
	Version         int                `json:"version"`
	RequestedTicks  uint64             `json:"requested_ticks"`
	FromTick        uint64             `json:"from_tick"`
	ToTick          uint64             `json:"to_tick"`
	Samples         int                `json:"samples"`
	Population      Range              `json:"population"`
	Genomes         Range              `json:"genomes"`
	Lineages        Range              `json:"lineages"`
	Energy          Range              `json:"energy"`
	Copies          uint64             `json:"copies"`
	Deaths          uint64             `json:"deaths"`
	NewLineages     int                `json:"new_lineages"`
	DiversityStart  Diversity          `json:"diversity_start"`
	DiversityEnd    Diversity          `json:"diversity_end"`
	AliveAges       Ages               `json:"alive_ages_at_end"`
	Lifetimes       Lifetimes          `json:"completed_lifetimes"`
	Flows           Flows              `json:"resource_flows"`
	StructuresStart Structures         `json:"structures_start"`
	StructuresEnd   Structures         `json:"structures_end"`
	LargestObserved int                `json:"largest_structure_observed"`
	Interactions    []Edge             `json:"interaction_graph"`
	Activity        []GenomeActivity   `json:"genome_activity"`
	Discoveries     []Discovery        `json:"new_genomes"`
	AbsentAtEnd     []string           `json:"initial_genomes_absent_at_end"`
	Dominant        []Genome           `json:"dominant_genomes_at_end"`
	Failures        map[string]uint64  `json:"failed_attempts"`
	RulesStart      RuleIdentity       `json:"rules_start"`
	RulesEnd        RuleIdentity       `json:"rules_end"`
	RuleEvents      []dsl.Event        `json:"rule_events"`
	Narrative       []string           `json:"narrative_ru"`
}

// Summarize uses complete reporting intervals, never interpolates counts or
// infers lifetimes from sampled survivors. The actual boundaries are explicit.
func Summarize(frames []Metrics, requested uint64) (WindowSummary, error) {
	r := WindowSummary{Version: 1, RequestedTicks: requested, Interactions: []Edge{}, Activity: []GenomeActivity{}, Discoveries: []Discovery{}, AbsentAtEnd: []string{}, Dominant: []Genome{}, RuleEvents: []dsl.Event{}, Narrative: []string{}, Failures: map[string]uint64{}}
	if requested == 0 || len(frames) < 2 {
		return r, fmt.Errorf("need a positive window and at least two frames")
	}
	for i, m := range frames {
		if i > 0 && m.Tick <= frames[i-1].Tick {
			return r, fmt.Errorf("telemetry ticks must increase")
		}
	}
	last := frames[len(frames)-1]
	cut := uint64(0)
	if last.Tick > requested {
		cut = last.Tick - requested
	}
	start := 0
	for i := 1; i < len(frames)-1 && frames[i].Tick <= cut; i++ {
		start = i
	}
	frames = frames[start:]
	variation, err := summarizeVariation(frames)
	if err != nil {
		return r, err
	}
	r.Variation = variation
	environment, err := summarizeEnvironment(frames)
	if err != nil {
		return r, err
	}
	r.Environment = environment
	first := frames[0]
	if first.Telemetry == nil || first.Telemetry.Version != 1 {
		return r, fmt.Errorf("event telemetry v1 is required; old metrics cannot recover exact lifetimes or interactions")
	}
	if len(first.Telemetry.SessionHash) != 64 {
		return r, fmt.Errorf("telemetry session identity is missing; record a new run")
	}
	r.Seed = first.Telemetry.Seed
	r.SessionHash = first.Telemetry.SessionHash
	r.FromTick = first.Tick
	r.ToTick = last.Tick
	r.Samples = len(frames)
	rangeOf := func(get func(Metrics) int64) Range {
		v := get(first)
		x := Range{Start: v, End: get(last), Min: v, Max: v}
		for _, m := range frames {
			x.Min = min(x.Min, get(m))
			x.Max = max(x.Max, get(m))
		}
		return x
	}
	r.Population = rangeOf(func(m Metrics) int64 { return int64(m.Entities) })
	r.Genomes = rangeOf(func(m Metrics) int64 { return int64(m.Genomes) })
	r.Lineages = rangeOf(func(m Metrics) int64 { return int64(m.Lineages) })
	r.Energy = rangeOf(func(m Metrics) int64 { return m.Energy })
	r.DiversityStart = first.Telemetry.Diversity
	r.StructuresStart = first.Telemetry.Structures
	r.RulesStart = first.Telemetry.ActiveRules
	edges := map[edgeKey]Edge{}
	activities := map[string]*GenomeActivity{}
	discoveries := map[string]Discovery{}
	hist := map[uint64]LifeBucket{}
	activity := func(hash string) *GenomeActivity {
		if activities[hash] == nil {
			activities[hash] = &GenomeActivity{Hash: hash}
		}
		return activities[hash]
	}
	for _, g := range first.ActiveGenomes {
		activity(g.Hash).Start = g.Count
	}
	r.Flows.DSL = map[string]uint64{}
	r.Lifetimes.Histogram = []LifeBucket{}
	for i, m := range frames {
		if m.Telemetry == nil || m.Telemetry.Version != 1 {
			return r, fmt.Errorf("missing event telemetry at tick %d", m.Tick)
		}
		t := m.Telemetry
		r.LargestObserved = max(r.LargestObserved, t.Structures.Largest)
		if i == 0 {
			continue
		}
		prev := frames[i-1]
		if err := validateInterval(prev, m); err != nil {
			return r, fmt.Errorf("tick %d: %w", m.Tick, err)
		}
		r.Copies += m.Copies - prev.Copies
		r.Deaths += m.Deaths - prev.Deaths
		for _, e := range t.Interactions {
			key := edgeKey{e.Kind, e.Source, e.Target}
			v := edges[key]
			v.Kind = e.Kind
			v.Source = e.Source
			v.Target = e.Target
			v.Count += e.Count
			v.Energy += e.Energy
			edges[key] = v
			if e.Kind == "copy" {
				activity(e.Source).Copies += e.Count
				activity(e.Target).Births += e.Count
			}
		}
		for _, d := range t.DeathsByGenome {
			if d.Hash != "" {
				activity(d.Hash).Deaths += d.Count
			}
		}
		for _, d := range t.Discoveries {
			if _, ok := discoveries[d.Hash]; ok {
				return r, fmt.Errorf("duplicate genome discovery %s", d.Hash)
			}
			discoveries[d.Hash] = d
		}
		a, b := &r.Lifetimes, t.Lifetimes
		if b.Count > 0 {
			if a.Count == 0 {
				a.Min = b.Min
			}
			a.Min = min(a.Min, b.Min)
			a.Max = max(a.Max, b.Max)
			a.Count += b.Count
			a.Sum += b.Sum
		}
		for _, bucket := range b.Histogram {
			v := hist[bucket.Min]
			v.Min = bucket.Min
			v.Max = bucket.Max
			v.Count += bucket.Count
			hist[bucket.Min] = v
		}
		addFlows(&r.Flows, t.Flows)
		r.RuleEvents = append(r.RuleEvents, t.RuleEvents...)
	}
	r.NewLineages = last.OriginsSeen - first.OriginsSeen
	r.DiversityEnd = last.Telemetry.Diversity
	r.AliveAges = last.Telemetry.AliveAges
	r.StructuresEnd = last.Telemetry.Structures
	r.RulesEnd = last.Telemetry.ActiveRules
	for _, g := range last.ActiveGenomes {
		activity(g.Hash).End = g.Count
		var previous Genome
		known := g.FirstTick >= r.FromTick
		for _, old := range first.ActiveGenomes {
			if old.Hash == g.Hash {
				previous = old
				known = true
				break
			}
		}
		if known {
			if g.Instructions < previous.Instructions || g.Moved < previous.Moved || g.Binds < previous.Binds || g.Converted[0] < previous.Converted[0] || g.Converted[1] < previous.Converted[1] {
				return r, fmt.Errorf("genome behavior counters decrease")
			}
			activity(g.Hash).Behavior = &Behavior{Instructions: g.Instructions - previous.Instructions, Moves: g.Moved - previous.Moved, Binds: g.Binds - previous.Binds, Converted: [2]int64{g.Converted[0] - previous.Converted[0], g.Converted[1] - previous.Converted[1]}}
		}
		r.Dominant = append(r.Dominant, g)
	}
	sort.Slice(r.Dominant, func(i, j int) bool {
		a, b := r.Dominant[i], r.Dominant[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.Hash < b.Hash
	})
	if len(r.Dominant) > 5 {
		r.Dominant = r.Dominant[:5]
	}
	for _, a := range activities {
		r.Activity = append(r.Activity, *a)
		if a.Start > 0 && a.End == 0 {
			r.AbsentAtEnd = append(r.AbsentAtEnd, a.Hash)
		}
	}
	sort.Slice(r.Activity, func(i, j int) bool { return r.Activity[i].Hash < r.Activity[j].Hash })
	sort.Strings(r.AbsentAtEnd)
	for _, d := range discoveries {
		r.Discoveries = append(r.Discoveries, d)
	}
	sort.Slice(r.Discoveries, func(i, j int) bool { return r.Discoveries[i].Hash < r.Discoveries[j].Hash })
	for _, e := range edges {
		r.Interactions = append(r.Interactions, e)
	}
	sortEdges(r.Interactions)
	for _, b := range hist {
		r.Lifetimes.Histogram = append(r.Lifetimes.Histogram, b)
	}
	sort.Slice(r.Lifetimes.Histogram, func(i, j int) bool { return r.Lifetimes.Histogram[i].Min < r.Lifetimes.Histogram[j].Min })
	if r.Lifetimes.Count > 0 {
		r.Lifetimes.Mean = float64(r.Lifetimes.Sum) / float64(r.Lifetimes.Count)
	}
	for name, pair := range failureCounters(first, last) {
		if pair[1] < pair[0] {
			return r, fmt.Errorf("decreasing failure counter")
		}
		r.Failures[name] = pair[1] - pair[0]
	}
	r.Narrative = describe(r)
	return r, nil
}

func addFlows(a *Flows, b Flows) {
	a.Injected += b.Injected
	a.Dissipated += b.Dissipated
	a.Absorbed += b.Absorbed
	a.Transferred += b.Transferred
	a.Taken += b.Taken
	a.Allocated += b.Allocated
	a.Charged += b.Charged
	for i := range a.Converted {
		a.Converted[i] += b.Converted[i]
	}
	for k, v := range b.DSL {
		a.DSL[k] += v
	}
}
func failureCounters(a, b Metrics) map[string][2]uint64 {
	return map[string][2]uint64{"space": {a.FailedSpace, b.FailedSpace}, "matter": {a.FailedMatter, b.FailedMatter}, "reserve": {a.FailedReserve, b.FailedReserve}, "limit": {a.FailedLimit, b.FailedLimit}, "absorb": {a.FailedAbsorb, b.FailedAbsorb}, "reaction": {a.FailedReaction, b.FailedReaction}, "instruction_energy": {a.InstructionStarved, b.InstructionStarved}}
}

func validateInterval(a, b Metrics) error {
	t := b.Telemetry
	if !t.Complete || t.FromTick != a.Tick || t.SessionStart != a.Telemetry.SessionStart || t.SessionHash != a.Telemetry.SessionHash || t.Seed != a.Telemetry.Seed {
		return fmt.Errorf("incomplete interval or mixed observation sessions")
	}
	if b.Copies < a.Copies || b.Deaths < a.Deaths || b.Absorbed < a.Absorbed || b.Transferred < a.Transferred || b.Allocations < a.Allocations || b.OriginsSeen < a.OriginsSeen {
		return fmt.Errorf("counters decrease")
	}
	if b.Deaths-a.Deaths != t.Lifetimes.Count {
		return fmt.Errorf("death/lifetime count mismatch")
	}
	if b.Energy-a.Energy != t.Flows.Injected-t.Flows.Dissipated || b.Matter != a.Matter || t.Pools.Field+t.Pools.Particle+t.Pools.Chemical+t.Pools.Signal != b.Energy {
		return fmt.Errorf("resource budget mismatch")
	}
	if t.Flows.Injected < 0 || t.Flows.Dissipated < 0 || t.Flows.Absorbed < 0 || t.Flows.Transferred < 0 || t.Flows.Taken < 0 || t.Flows.Charged < 0 || t.Flows.Converted[0] < 0 || t.Flows.Converted[1] < 0 {
		return fmt.Errorf("negative flow")
	}
	if t.Flows.Absorbed != b.Absorbed-a.Absorbed || t.Flows.Transferred != b.Transferred-a.Transferred || t.Flows.Allocated != int64(b.Allocations-a.Allocations)*12 {
		return fmt.Errorf("flow counters mismatch")
	}
	for i := range b.Converted {
		if b.Converted[i] < a.Converted[i] || t.Flows.Converted[i] != b.Converted[i]-a.Converted[i] {
			return fmt.Errorf("reaction counters mismatch")
		}
	}
	seen := map[edgeKey]bool{}
	var copies uint64
	var transferred, taken, allocated int64
	balance := map[string]int64{}
	for _, g := range a.ActiveGenomes {
		if g.Count < 1 {
			return fmt.Errorf("invalid genome count")
		}
		balance[g.Hash] += int64(g.Count)
	}
	for _, e := range t.Interactions {
		key := edgeKey{e.Kind, e.Source, e.Target}
		if seen[key] || e.Count == 0 || e.Energy < 0 {
			return fmt.Errorf("invalid interaction edge")
		}
		seen[key] = true
		switch e.Kind {
		case "copy":
			copies += e.Count
			balance[e.Target] += int64(e.Count)
			if e.Energy != 0 || e.Source == "" || e.Target == "" {
				return fmt.Errorf("invalid copy edge")
			}
		case "transfer":
			transferred += e.Energy
		case "take":
			taken += e.Energy
		case "allocate":
			allocated += e.Energy
			if e.Energy != int64(e.Count)*12 {
				return fmt.Errorf("allocation reserve mismatch")
			}
		case "bind", "unbind":
			if e.Energy != 0 {
				return fmt.Errorf("binding has energy flow")
			}
		default:
			return fmt.Errorf("unknown interaction kind")
		}
	}
	if copies != b.Copies-a.Copies || transferred != t.Flows.Transferred || taken != t.Flows.Taken || allocated != t.Flows.Allocated {
		return fmt.Errorf("interaction graph does not reconcile with accounting")
	}
	var deaths, buckets uint64
	seenDeaths := map[string]bool{}
	for _, d := range t.DeathsByGenome {
		if seenDeaths[d.Hash] {
			return fmt.Errorf("duplicate genome death counter")
		}
		seenDeaths[d.Hash] = true
		deaths += d.Count
		if d.Hash != "" {
			balance[d.Hash] -= int64(d.Count)
		}
	}
	for _, g := range b.ActiveGenomes {
		balance[g.Hash] -= int64(g.Count)
	}
	for _, v := range balance {
		if v != 0 {
			return fmt.Errorf("genome birth/death balance mismatch")
		}
	}
	for _, bucket := range t.Lifetimes.Histogram {
		if bucket.Min > bucket.Max || bucket.Count == 0 {
			return fmt.Errorf("invalid lifetime bucket")
		}
		buckets += bucket.Count
	}
	if deaths != t.Lifetimes.Count || buckets != deaths {
		return fmt.Errorf("lifetime histogram does not reconcile")
	}
	for _, d := range t.Discoveries {
		if d.Hash == "" || d.FirstTick < a.Tick || d.FirstTick >= b.Tick {
			return fmt.Errorf("invalid discovery boundary")
		}
	}
	for _, e := range t.RuleEvents {
		if e.Tick < a.Tick || e.Tick > b.Tick {
			return fmt.Errorf("rule event outside interval")
		}
	}
	return nil
}

func describe(r WindowSummary) []string {
	out := []string{fmt.Sprintf("Тики %d–%d: %d наблюдений; фактическое окно %d тиков при запросе %d.", r.FromTick, r.ToTick, r.Samples, r.ToTick-r.FromTick, r.RequestedTicks),
		fmt.Sprintf("Популяция: %d → %d; на отчётных кадрах от %d до %d. Копирований кода: %d, смертей частиц: %d.", r.Population.Start, r.Population.End, r.Population.Min, r.Population.Max, r.Copies, r.Deaths),
		fmt.Sprintf("Геномов: %d → %d; эффективное разнообразие: %.2f → %.2f. Новых геномов за окно: %d; исходных геномов, отсутствующих в конце: %d.", r.Genomes.Start, r.Genomes.End, r.DiversityStart.EffectiveGenomes, r.DiversityEnd.EffectiveGenomes, len(r.Discoveries), len(r.AbsentAtEnd))}
	if len(r.Dominant) > 0 {
		g := r.Dominant[0]
		var copies uint64
		for _, a := range r.Activity {
			if a.Hash == g.Hash {
				copies = a.Copies
			}
		}
		out = append(out, fmt.Sprintf("В конце доминирует %s: %d частиц, %.1f%% исполняемых частиц; за окно эта линия выполнила %d копирований.", g.Hash, g.Count, 100*g.Frequency, copies))
		for _, a := range r.Activity {
			if a.Hash == g.Hash && a.Behavior != nil {
				b := a.Behavior
				out = append(out, fmt.Sprintf("Действия доминирующей линии за окно: перемещений %d, новых связей %d, превращений по ID 0/1: %d/%d.", b.Moves, b.Binds, b.Converted[0], b.Converted[1]))
			}
		}
	}
	if r.Lifetimes.Count > 0 {
		out = append(out, fmt.Sprintf("Среди умерших в окне среднее полное время жизни %.2f тика, минимум %d, максимум %d. Возраст живых в конце: P50=%d, P90=%d тиков.", r.Lifetimes.Mean, r.Lifetimes.Min, r.Lifetimes.Max, r.AliveAges.Median, r.AliveAges.P90))
	} else {
		out = append(out, "Смертей в окне не наблюдалось; завершённые времена жизни недоступны.")
	}
	out = append(out, fmt.Sprintf("Связанных структур в конце: %d; связанных частиц: %d. Самая большая структура на отчётных кадрах: %d частиц; структур с несколькими геномами в конце: %d.", r.StructuresEnd.LinkedComponents, r.StructuresEnd.LinkedParticles, r.LargestObserved, r.StructuresEnd.MixedGenomeComponents),
		fmt.Sprintf("Энергия: поступило %d, рассеяно %d, запас изменился на %+d. Поглощено из поля %d; передано TRANSFER %d, отобрано TAKE %d, выделено в стартовый резерв %d.", r.Flows.Injected, r.Flows.Dissipated, r.Energy.End-r.Energy.Start, r.Flows.Absorbed, r.Flows.Transferred, r.Flows.Taken, r.Flows.Allocated),
		fmt.Sprintf("Отказы: нет места %d, материи %d, стартового резерва %d; недостаточно энергии инструкции %d; неуспешных реакций %d.", r.Failures["space"], r.Failures["matter"], r.Failures["reserve"], r.Failures["instruction_energy"], r.Failures["reaction"]))
	if len(r.RuleEvents) > 0 {
		out = append(out, fmt.Sprintf("За окно зарегистрировано смен/откатов правил: %d. Семантику реакций следует сопоставлять с версией модуля.", len(r.RuleEvents)))
	}
	out = append(out, "Граф фиксирует действия и прямые передачи энергии между геномами; обмен через химические поля не атрибутирован отдельным линиям. Сосуществование и связи сами по себе не доказывают взаимозависимость.")
	return out
}
