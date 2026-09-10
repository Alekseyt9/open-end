package observer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
)

// DynamicsConfig describes a local, sampled heuristic, not adaptive fitness.
type DynamicsConfig struct {
	MinTicks           uint64  `json:"min_ticks"`
	PlateauTolerance   float64 `json:"plateau_relative_tolerance"`
	MonocultureShare   float64 `json:"monoculture_share"`
	RepeatThreshold    float64 `json:"repeat_threshold"`
	BehaviorResolution int     `json:"behavior_bins_per_octave"`
}

func DefaultDynamicsConfig() DynamicsConfig {
	return DynamicsConfig{10000, 0.1, 0.9, 0.75, 4}
}

type Span struct {
	Min           float64 `json:"min"`
	Max           float64 `json:"max"`
	RelativeRange float64 `json:"relative_range"`
}

type BehaviorSignature struct {
	FromTick uint64    `json:"from_tick"`
	ToTick   uint64    `json:"to_tick"`
	Hash     string    `json:"sha256"`
	Rates    []float64 `json:"rates_per_1000_particle_ticks"`
	Bins     []int     `json:"quantized_rates"`
}

type Dynamics struct {
	Version                   int                 `json:"version"`
	Status                    string              `json:"status"`
	FromTick                  uint64              `json:"from_tick"`
	ToTick                    uint64              `json:"to_tick"`
	Config                    DynamicsConfig      `json:"config"`
	Diversity                 Span                `json:"effective_diversity"`
	Population                Span                `json:"population"`
	LargestStructure          Span                `json:"largest_structure"`
	StructureDistance         float64             `json:"max_size_distribution_total_variation"`
	DiversityPlateau          bool                `json:"diversity_plateau"`
	StructuralPlateau         bool                `json:"structural_plateau"`
	Monoculture               bool                `json:"persistent_same_genome_monoculture"`
	RepeatedFraction          float64             `json:"repeated_behavior_fraction"`
	PersistentNewBehavior     bool                `json:"persistent_new_behavior"`
	BehaviorSeparation        float64             `json:"min_persistent_behavior_separation_octaves"`
	PersistentStructureGrowth bool                `json:"persistent_structure_growth"`
	NewGenomes                int                 `json:"new_genomes_not_used_as_novelty"`
	Features                  []string            `json:"behavior_features"`
	Behavior                  []BehaviorSignature `json:"behavior_blocks"`
	Reasons                   []string            `json:"reasons_ru"`
}

// DetectDynamics validates telemetry through Summarize, then compares four
// consecutive time blocks. The first two form the local reference; both later
// blocks must support a novel signature/growth before declaring development.
// No state is retained across calls and no simulation state is changed.
func DetectDynamics(frames []Metrics, requested uint64, cfg DynamicsConfig) (Dynamics, error) {
	r := Dynamics{Version: 1, Status: "insufficient_history", Config: cfg,
		Behavior: []BehaviorSignature{}, Reasons: []string{},
		Features: []string{"allocate", "copy", "transfer", "take", "bind", "unbind", "absorbed_energy", "converted_x", "converted_y", "charged_z", "dsl_units"}}
	validFraction := func(x float64) bool { return !math.IsNaN(x) && x > 0 && x <= 1 }
	if cfg.MinTicks == 0 || !validFraction(cfg.PlateauTolerance) || !validFraction(cfg.MonocultureShare) || !validFraction(cfg.RepeatThreshold) || cfg.BehaviorResolution < 1 || cfg.BehaviorResolution > 64 {
		return r, fmt.Errorf("invalid dynamics configuration")
	}
	summary, err := Summarize(frames, requested)
	if err != nil {
		return r, err
	}
	for frames[0].Tick < summary.FromTick {
		frames = frames[1:]
	}
	r.FromTick, r.ToTick = summary.FromTick, summary.ToTick
	r.NewGenomes = len(summary.Discoveries)
	first, last := frames[0], frames[len(frames)-1]
	if last.Executable == 0 {
		r.Status = "extinct"
		r.Reasons = append(r.Reasons, "Исполняемых частиц в конце окна нет; это вымирание, а не плато живой популяции.")
		return r, nil
	}
	duration := r.ToTick - r.FromTick
	if duration < cfg.MinTicks || len(frames) < 5 {
		r.Reasons = append(r.Reasons, fmt.Sprintf("Недостаточно истории: нужно минимум %d тиков и четыре интервала наблюдения.", cfg.MinTicks))
		return r, nil
	}
	// Select real boundaries at or before each quarter. Sparse/uneven reporting
	// may not cover all quarters; never fabricate events or interpolate them.
	boundaries := []int{0}
	for quarter := uint64(1); quarter < 4; quarter++ {
		target := r.FromTick + duration/4*quarter + duration%4*quarter/4
		index := boundaries[len(boundaries)-1]
		for index+1 < len(frames)-1 && frames[index+1].Tick <= target {
			index++
		}
		if index == boundaries[len(boundaries)-1] {
			r.Reasons = append(r.Reasons, "Слишком редкие кадры: нет отдельного интервала в каждой четверти окна.")
			return r, nil
		}
		boundaries = append(boundaries, index)
	}
	boundaries = append(boundaries, len(frames)-1)
	r.Diversity = sampledSpan(frames, func(m Metrics) float64 { return m.Telemetry.Diversity.EffectiveGenomes })
	r.Population = sampledSpan(frames, func(m Metrics) float64 { return float64(m.Entities) })
	r.LargestStructure = sampledSpan(frames, func(m Metrics) float64 { return float64(m.Telemetry.Structures.Largest) })
	r.Monoculture = true
	dominant := dominantHash(first)
	baselineSizes := sizeDistribution(first)
	for _, m := range frames {
		if m.Telemetry.Diversity.DominantShare < cfg.MonocultureShare || dominant == "" || dominantHash(m) != dominant {
			r.Monoculture = false
		}
		r.StructureDistance = max(r.StructureDistance, distributionDistance(baselineSizes, sizeDistribution(m)))
	}
	r.DiversityPlateau = r.Diversity.RelativeRange <= cfg.PlateauTolerance
	r.StructuralPlateau = r.LargestStructure.RelativeRange <= cfg.PlateauTolerance && r.StructureDistance <= cfg.PlateauTolerance
	seen := map[string]bool{}
	repeated := 0
	for i := 0; i < 4; i++ {
		block := frames[boundaries[i] : boundaries[i+1]+1]
		s, err := Summarize(block, block[len(block)-1].Tick-block[0].Tick)
		if err != nil {
			return r, err
		}
		b := behaviorSignature(block, s, cfg.BehaviorResolution)
		if seen[b.Hash] {
			repeated++
		}
		seen[b.Hash] = true
		r.Behavior = append(r.Behavior, b)
	}
	r.RepeatedFraction = float64(repeated) / 3
	a, b, c, d := r.Behavior[0].Hash, r.Behavior[1].Hash, r.Behavior[2].Hash, r.Behavior[3].Hash
	// Merely crossing a quantization boundary can change a hash while rates
	// barely move. Require a full bin of raw log-rate separation from BOTH
	// reference blocks in BOTH confirmation blocks.
	r.BehaviorSeparation = math.Inf(1)
	for _, recent := range r.Behavior[2:] {
		for _, old := range r.Behavior[:2] {
			distance := 0.0
			for i := range recent.Rates {
				distance = max(distance, math.Abs(math.Log2(1+recent.Rates[i])-math.Log2(1+old.Rates[i])))
			}
			r.BehaviorSeparation = min(r.BehaviorSeparation, distance)
		}
	}
	r.PersistentNewBehavior = c == d && c != a && c != b && r.BehaviorSeparation >= 1/float64(cfg.BehaviorResolution)
	baselineMax := 0
	for _, m := range frames[:boundaries[2]+1] {
		baselineMax = max(baselineMax, m.Telemetry.Structures.Largest)
	}
	r.PersistentStructureGrowth = true
	for _, m := range frames[boundaries[2]+1:] {
		if float64(m.Telemetry.Structures.Largest-baselineMax) <= max(2, float64(baselineMax)*cfg.PlateauTolerance) {
			r.PersistentStructureGrowth = false
		}
	}
	r.Reasons = append(r.Reasons,
		fmt.Sprintf("Размах эффективного разнообразия %.1f%%; плато=%t. Устойчивая монокультура=%t.", 100*r.Diversity.RelativeRange, r.DiversityPlateau, r.Monoculture),
		fmt.Sprintf("Размах крупнейшей структуры %.1f%%; изменение распределения размеров %.3f; плато=%t.", 100*r.LargestStructure.RelativeRange, r.StructureDistance, r.StructuralPlateau),
		fmt.Sprintf("Повторение поведенческих хешей: %.1f%%. Устойчивое новое поведение=%t; устойчивый рост структур=%t.", 100*r.RepeatedFraction, r.PersistentNewBehavior, r.PersistentStructureGrowth))
	switch {
	case len(summary.RuleEvents) > 0 || summary.RulesStart != summary.RulesEnd:
		r.Status = "rule_change"
		r.Reasons = append(r.Reasons, "В окне менялись правила; различия не приписываются внутренней эволюции. Нужна история после смены правил.")
	case r.DiversityPlateau && r.StructuralPlateau && r.RepeatedFraction >= cfg.RepeatThreshold && r.Population.RelativeRange <= cfg.PlateauTolerance:
		r.Status = "stagnating"
		r.Reasons = append(r.Reasons, "Мир застрял по наблюдаемым признакам: плато разнообразия, численности и структур сопровождается повторением поведения.")
	case r.PersistentNewBehavior || r.PersistentStructureGrowth:
		r.Status = "developing"
		r.Reasons = append(r.Reasons, "Мир развивается по эвристике: изменение поведения или рост структур удерживается во второй половине окна. Адаптивная ценность ещё не проверена.")
	default:
		r.Status = "mixed"
		r.Reasons = append(r.Reasons, "Динамика неоднозначна: признаков недостаточно для устойчивой новизны или совместного плато.")
	}
	r.Reasons = append(r.Reasons, fmt.Sprintf("Новых геномов: %d; их появление само по себе не считается поведенческой новизной.", r.NewGenomes))
	return r, nil
}

func sampledSpan(frames []Metrics, get func(Metrics) float64) Span {
	r := Span{Min: get(frames[0]), Max: get(frames[0])}
	for _, m := range frames {
		r.Min = min(r.Min, get(m))
		r.Max = max(r.Max, get(m))
	}
	r.RelativeRange = (r.Max - r.Min) / max(1, r.Min)
	return r
}

func dominantHash(m Metrics) string {
	hash, count := "", 0
	for _, g := range m.ActiveGenomes {
		if g.Count > count || (g.Count == count && g.Hash < hash) {
			hash, count = g.Hash, g.Count
		}
	}
	return hash
}

// Particle-weighted size buckets: 1, 2–3, 4–7, ... . This is deliberately a
// size-distribution proxy; it does not establish topology or individuality.
func sizeDistribution(m Metrics) [64]float64 {
	var result [64]float64
	for _, s := range m.Telemetry.Structures.Sizes {
		bucket := 0
		for n := s.Size; n > 1; n /= 2 {
			bucket++
		}
		result[bucket] += float64(s.Size) * float64(s.Count) / float64(max(1, m.Entities))
	}
	return result
}

func distributionDistance(a, b [64]float64) float64 {
	v := 0.0
	for i := range a {
		v += math.Abs(a[i] - b[i])
	}
	return v / 2
}

func behaviorSignature(frames []Metrics, s WindowSummary, resolution int) BehaviorSignature {
	r := BehaviorSignature{FromTick: s.FromTick, ToTick: s.ToTick, Rates: make([]float64, 11), Bins: make([]int, 11)}
	// Exposure is a trapezoidal estimate from reporting frames; exact events in
	// the numerator are retained even when particles die between observations.
	exposure := 0.0
	for i := 1; i < len(frames); i++ {
		exposure += (float64(frames[i-1].Entities) + float64(frames[i].Entities)) / 2 * float64(frames[i].Tick-frames[i-1].Tick)
	}
	indices := map[string]int{"allocate": 0, "copy": 1, "transfer": 2, "take": 3, "bind": 4, "unbind": 5}
	for _, e := range s.Interactions {
		index := indices[e.Kind]
		r.Rates[index] += float64(e.Count)
	}
	r.Rates[6] = float64(s.Flows.Absorbed)
	r.Rates[7], r.Rates[8], r.Rates[9] = float64(s.Flows.Converted[0]), float64(s.Flows.Converted[1]), float64(s.Flows.Charged)
	// Sum integers before conversion, independent of DSL map iteration order.
	var units uint64
	for _, v := range s.Flows.DSL {
		units += v
	}
	r.Rates[10] = float64(units)
	for i := range r.Rates {
		r.Rates[i] *= 1000 / max(1, exposure)
		r.Bins[i] = int(math.Round(math.Log2(1+r.Rates[i]) * float64(resolution)))
	}
	encoded, _ := json.Marshal(struct {
		Version, Resolution int
		Bins                []int
	}{1, resolution, r.Bins})
	r.Hash = fmt.Sprintf("%x", sha256.Sum256(encoded))
	return r
}
