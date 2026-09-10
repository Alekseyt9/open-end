package adaptivity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"open-end/internal/world"
	"sort"
)

// FunctionalDiversity groups live genomes by coarse recent action profiles.
func FunctionalDiversity(source, end *world.World) (float64, []float64) {
	population := map[string]int{}
	for _, p := range end.Particles {
		if len(p.Code) > 0 {
			population[p.Genome]++
		}
	}
	groups := map[string]int{}
	descriptor := make([]float64, 7)
	total := 0
	keys := make([]string, 0, len(population))
	for h := range population {
		keys = append(keys, h)
	}
	sort.Strings(keys)
	for _, h := range keys {
		n := population[h]
		a := source.Genomes[h]
		if a == nil {
			a = &world.GenomeRecord{}
		}
		b := end.Genomes[h]
		instructions := b.Instructions - a.Instructions
		if instructions == 0 {
			continue
		}
		values := []float64{float64(b.Moved - a.Moved), float64(b.Copies - a.Copies), float64(b.Absorbed - a.Absorbed), float64(b.Transferred - a.Transferred), float64(b.Taken - a.Taken), float64(b.Binds - a.Binds), float64(b.Converted[0] + b.Converted[1] - a.Converted[0] - a.Converted[1])}
		bins := make([]int, len(values))
		for i, v := range values {
			rate := math.Log2(1 + 1000*v/float64(instructions))
			bins[i] = int(math.Round(2 * rate))
			descriptor[i] += float64(n) * rate
		}
		encoded, _ := json.Marshal(bins)
		groups[string(encoded)] += n
		total += n
	}
	if total == 0 {
		return 0, descriptor
	}
	entropy := 0.0
	ordered := make([]string, 0, len(groups))
	for k := range groups {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)
	for _, k := range ordered {
		p := float64(groups[k]) / float64(total)
		entropy -= p * math.Log(p)
	}
	for i := range descriptor {
		descriptor[i] /= float64(total)
	}
	return math.Exp(entropy), descriptor
}

type Candidate struct {
	Seed                uint64    `json:"seed"`
	SourceHash          string    `json:"source_sha256"`
	Selection           Score     `json:"selection"`
	Validation          Score     `json:"validation"`
	FunctionalDiversity float64   `json:"functional_diversity"`
	Descriptor          []float64 `json:"behavior_descriptor"`
	Probes              []Probe   `json:"probes"`
	Selected            bool      `json:"selected"`
	Reason              string    `json:"selection_reason,omitempty"`
}

// Select uses no reserved validation results. Up to half the slots reward
// positive adaptive scores; one explores; remaining slots spread behaviors.
func Select(candidates []Candidate, slots int, seed uint64) error {
	if slots < 1 || slots > len(candidates) {
		return fmt.Errorf("invalid selection slots")
	}
	seen := map[string]bool{}
	for i := range candidates {
		c := &candidates[i]
		if math.IsNaN(c.Selection.AdaptiveProxy) || math.IsInf(c.Selection.AdaptiveProxy, 0) || c.Selection.AdaptiveProxy < 0 || c.Selection.AdaptiveProxy > 100 || math.IsNaN(c.FunctionalDiversity) || math.IsInf(c.FunctionalDiversity, 0) || c.FunctionalDiversity < 0 {
			return fmt.Errorf("invalid candidate score")
		}
		for _, v := range c.Descriptor {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
				return fmt.Errorf("invalid descriptor")
			}
		}
		if seen[c.SourceHash] || len(c.SourceHash) != 64 || len(c.Descriptor) != 7 {
			return fmt.Errorf("invalid candidate")
		}
		seen[c.SourceHash] = true
		c.Selected = false
		c.Reason = ""
	}
	indices := make([]int, len(candidates))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool {
		a, b := candidates[indices[i]], candidates[indices[j]]
		if a.Selection.AdaptiveProxy != b.Selection.AdaptiveProxy {
			return a.Selection.AdaptiveProxy > b.Selection.AdaptiveProxy
		}
		return a.SourceHash < b.SourceHash
	})
	selected := 0
	for _, i := range indices {
		if selected >= slots/2 || candidates[i].Selection.AdaptiveProxy <= 0 {
			break
		}
		candidates[i].Selected = true
		candidates[i].Reason = "adaptive_proxy"
		selected++
	}
	exploreSlots := 1
	if slots == 1 {
		exploreSlots = 0
	}
	for selected < slots-exploreSlots {
		best := -1
		bestDistance := -1.0
		for i, c := range candidates {
			if c.Selected {
				continue
			}
			distance := math.Inf(1)
			hasReference := false
			for _, r := range candidates {
				if !r.Selected {
					continue
				}
				hasReference = true
				d := 0.0
				for j, v := range c.Descriptor {
					d += (v - r.Descriptor[j]) * (v - r.Descriptor[j])
				}
				distance = math.Min(distance, d)
			}
			if !hasReference {
				distance = c.FunctionalDiversity
			}
			if best < 0 || distance > bestDistance || distance == bestDistance && c.SourceHash < candidates[best].SourceHash {
				best = i
				bestDistance = distance
			}
		}
		candidates[best].Selected = true
		candidates[best].Reason = "behavioral_diversity"
		selected++
	}
	if selected < slots {
		best := -1
		rank := ""
		for i, c := range candidates {
			if c.Selected {
				continue
			}
			h := sha256.Sum256([]byte(fmt.Sprintf("%d/%s", seed, c.SourceHash)))
			s := hex.EncodeToString(h[:])
			if best < 0 || s < rank {
				best = i
				rank = s
			}
		}
		candidates[best].Selected = true
		candidates[best].Reason = "exploration"
	}
	return nil
}
