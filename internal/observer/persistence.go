package observer

import (
	"fmt"
	"sort"
)

// PersistenceOptions are observer thresholds, never fitness or physics rules.
type PersistenceOptions struct {
	Windows   int    `json:"windows"`
	MinCount  int    `json:"min_count"`
	MinCopies uint64 `json:"min_copies_per_window"`
}

type PersistentGenome struct {
	Hash           string   `json:"hash"`
	Count          int      `json:"count"`
	Frequency      float64  `json:"frequency"`
	MinCount       int      `json:"min_count_observed"`
	Copies         uint64   `json:"copies_in_window"`
	Instructions   uint64   `json:"instructions_in_window"`
	Moved          uint64   `json:"moves_in_window"`
	Binds          uint64   `json:"binds_in_window"`
	Transferred    int64    `json:"energy_transferred_in_window"`
	Taken          int64    `json:"energy_taken_in_window"`
	Converted      [2]int64 `json:"converted_in_window"`
	Reaction1Share float64  `json:"reaction_1_share"`
	Profile        string   `json:"observed_profile"`
}

type Persistence struct {
	FromTick                  uint64             `json:"from_tick"`
	ToTick                    uint64             `json:"to_tick"`
	Options                   PersistenceOptions `json:"thresholds"`
	Genomes                   []PersistentGenome `json:"persistent_genomes"`
	Profiles                  []string           `json:"observed_profiles"`
	HasComplementaryReactions bool               `json:"has_complementary_reaction_candidates"`
}

// Persistent requires presence at every boundary and offspring in EVERY recent
// window. Missing observations are censored, not treated as zero counters.
// This avoids mistaking old accumulated copies or immortal survivors for growth.
func Persistent(frames []Metrics, options PersistenceOptions) (Persistence, error) {
	r := Persistence{Options: options, Genomes: []PersistentGenome{}, Profiles: []string{}}
	if options.Windows < 1 || options.MinCount < 1 || options.MinCopies < 1 {
		return r, fmt.Errorf("persistence thresholds must be positive")
	}
	if len(frames) < options.Windows+1 {
		return r, fmt.Errorf("need at least %d metric frames", options.Windows+1)
	}
	frames = frames[len(frames)-options.Windows-1:]
	lookups := make([]map[string]Genome, len(frames))
	for j, m := range frames {
		if j > 0 && m.Tick <= frames[j-1].Tick {
			return r, fmt.Errorf("ticks must increase")
		}
		lookups[j] = make(map[string]Genome)
		for _, g := range m.ActiveGenomes {
			if _, found := lookups[j][g.Hash]; found {
				return r, fmt.Errorf("duplicate genome at tick %d", m.Tick)
			}
			lookups[j][g.Hash] = g
		}
	}
	r.FromTick = frames[0].Tick
	r.ToTick = frames[len(frames)-1].Tick
	profiles := make(map[string]bool)
	has0, has1 := false, false
	for hash, last := range lookups[len(lookups)-1] {
		g := PersistentGenome{Hash: hash, Count: last.Count, Frequency: last.Frequency, MinCount: last.Count}
		valid := true
		for j := range lookups {
			current, ok := lookups[j][hash]
			if !ok || current.Count < options.MinCount {
				valid = false
				break
			}
			g.MinCount = min(g.MinCount, current.Count)
			if j == 0 {
				continue
			}
			previous := lookups[j-1][hash]
			if current.Copies < previous.Copies || current.Instructions < previous.Instructions || current.Moved < previous.Moved || current.Binds < previous.Binds || current.Transferred < previous.Transferred || current.Taken < previous.Taken || current.Converted[0] < previous.Converted[0] || current.Converted[1] < previous.Converted[1] {
				return r, fmt.Errorf("counters decrease for %s", hash)
			}
			copies := current.Copies - previous.Copies
			if copies < options.MinCopies {
				valid = false
				break
			}
			g.Copies += copies
			g.Instructions += current.Instructions - previous.Instructions
			g.Moved += current.Moved - previous.Moved
			g.Binds += current.Binds - previous.Binds
			g.Transferred += current.Transferred - previous.Transferred
			g.Taken += current.Taken - previous.Taken
			for i := range g.Converted {
				g.Converted[i] += current.Converted[i] - previous.Converted[i]
			}
		}
		if !valid {
			continue
		}
		// Descriptive bins, not organism classes. Report raw counts alongside bins.
		binding := "few-binds"
		if float64(g.Binds)/float64(g.Copies) >= 0.25 {
			binding = "frequent-binds"
		}
		movement := "few-moves"
		if g.Instructions > 0 && float64(g.Moved)/float64(g.Instructions) >= 0.01 {
			movement = "frequent-moves"
		}
		reaction := "low-conversion"
		total := g.Converted[0] + g.Converted[1]
		if total >= 100 {
			g.Reaction1Share = float64(g.Converted[1]) / float64(total)
			reaction = "mixed-reactions"
			if g.Reaction1Share <= 0.1 {
				reaction = "reaction-0-dominant"
				has0 = true
			}
			if g.Reaction1Share >= 0.9 {
				reaction = "reaction-1-dominant"
				has1 = true
			}
		}
		g.Profile = binding + "/" + movement + "/" + reaction
		profiles[g.Profile] = true
		r.Genomes = append(r.Genomes, g)
	}
	r.HasComplementaryReactions = has0 && has1
	for p := range profiles {
		r.Profiles = append(r.Profiles, p)
	}
	sort.Strings(r.Profiles)
	sort.Slice(r.Genomes, func(i, j int) bool { return r.Genomes[i].Hash < r.Genomes[j].Hash })
	return r, nil
}
