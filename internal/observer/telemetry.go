package observer

import (
	"math"
	"open-end/internal/dsl"
	"open-end/internal/world"
	"sort"
)

type Diversity struct {
	Shannon          float64 `json:"shannon_nats"`
	EffectiveGenomes float64 `json:"effective_genomes"`
	InverseSimpson   float64 `json:"inverse_simpson"`
	DominantShare    float64 `json:"dominant_share"`
	LineageShannon   float64 `json:"lineage_shannon_nats"`
}
type Ages struct {
	Count  int    `json:"count"`
	Median uint64 `json:"median_ticks"`
	P90    uint64 `json:"p90_ticks"`
	Max    uint64 `json:"max_ticks"`
}
type SizeCount struct {
	Size  int `json:"size"`
	Count int `json:"count"`
}
type Binding struct {
	A     string `json:"a"`
	B     string `json:"b"`
	Count int    `json:"count"`
}
type Structures struct {
	Components            int         `json:"components"`
	LinkedComponents      int         `json:"linked_components"`
	LinkedParticles       int         `json:"linked_particles"`
	MixedGenomeComponents int         `json:"mixed_genome_components"`
	Largest               int         `json:"largest"`
	Sizes                 []SizeCount `json:"size_histogram"`
	Bindings              []Binding   `json:"genome_binding_graph"`
}
type Pools struct {
	Signal   int64 `json:"signal_energy,omitempty"`
	Field    int64 `json:"field_energy"`
	Particle int64 `json:"particle_energy"`
	Chemical int64 `json:"chemical_energy"`
}
type Flows struct {
	Injected    int64             `json:"injected_energy"`
	Dissipated  int64             `json:"dissipated_energy"`
	Absorbed    int64             `json:"absorbed_energy"`
	Transferred int64             `json:"transferred_energy"`
	Taken       int64             `json:"taken_energy"`
	Allocated   int64             `json:"allocation_reserve_energy"`
	Charged     int64             `json:"charged_z_to_x_units"`
	Converted   [2]int64          `json:"converted_by_id"`
	DSL         map[string]uint64 `json:"dsl_units_by_hash_and_name"`
}
type Edge struct {
	Kind   string `json:"kind"`
	Source string `json:"source_genome"`
	Target string `json:"target_genome"`
	Count  uint64 `json:"events"`
	Energy int64  `json:"energy"`
}
type LifeBucket struct {
	Min   uint64 `json:"min_ticks"`
	Max   uint64 `json:"max_ticks"`
	Count uint64 `json:"count"`
}
type Lifetimes struct {
	Count     uint64       `json:"deaths"`
	Sum       uint64       `json:"sum_lifetime_ticks"`
	Mean      float64      `json:"mean_lifetime_ticks"`
	Min       uint64       `json:"min_lifetime_ticks"`
	Max       uint64       `json:"max_lifetime_ticks"`
	Histogram []LifeBucket `json:"histogram"`
}
type GenomeDeaths struct {
	Hash  string `json:"hash"`
	Count uint64 `json:"count"`
}
type Discovery struct {
	Hash      string `json:"hash"`
	Parent    string `json:"parent_genome"`
	FirstTick uint64 `json:"first_tick"`
}
type RuleIdentity struct {
	Version string `json:"version"`
	Hash    string `json:"sha256"`
}
type Telemetry struct {
	Collectives    *CollectiveFrame `json:"collectives,omitempty"`
	Seed           uint64           `json:"seed"`
	SessionHash    string           `json:"session_initial_world_sha256"`
	Version        int              `json:"version"`
	SessionStart   uint64           `json:"session_start_tick"`
	FromTick       uint64           `json:"from_tick"`
	Complete       bool             `json:"complete_interval"`
	Diversity      Diversity        `json:"diversity"`
	AliveAges      Ages             `json:"alive_ages"`
	Structures     Structures       `json:"structures"`
	Pools          Pools            `json:"energy_pools"`
	Flows          Flows            `json:"resource_flows"`
	Lifetimes      Lifetimes        `json:"completed_lifetimes"`
	DeathsByGenome []GenomeDeaths   `json:"deaths_by_genome"`
	Interactions   []Edge           `json:"interaction_graph"`
	Discoveries    []Discovery      `json:"new_genomes"`
	ActiveRules    RuleIdentity     `json:"active_rules"`
	RuleEvents     []dsl.Event      `json:"rule_events"`
}

func instantaneous(w *world.World, m Metrics, t *Telemetry) {
	squared := 0.0
	for _, g := range m.ActiveGenomes {
		p := g.Frequency
		t.Diversity.Shannon -= p * math.Log(p)
		squared += p * p
		t.Diversity.DominantShare = max(t.Diversity.DominantShare, p)
	}
	if m.Executable > 0 {
		t.Diversity.EffectiveGenomes = math.Exp(t.Diversity.Shannon)
		t.Diversity.InverseSimpson = 1 / squared
	}
	for _, l := range m.ActiveLineages {
		p := float64(l.Count) / float64(m.Executable)
		t.Diversity.LineageShannon -= p * math.Log(p)
	}
	ages := make([]uint64, 0, len(w.Particles))
	for _, p := range w.Particles {
		ages = append(ages, w.Tick-p.Created)
		t.Pools.Particle += int64(p.Energy)
	}
	sort.Slice(ages, func(i, j int) bool { return ages[i] < ages[j] })
	if len(ages) > 0 {
		t.AliveAges = Ages{len(ages), ages[(len(ages)-1)/2], ages[(9*len(ages)-1)/10], ages[len(ages)-1]}
	}
	for _, c := range w.Cells {
		t.Pools.Field += int64(c.Energy)
		t.Pools.Signal += int64(c.Signal)
		t.Pools.Chemical += int64(8*c.Chemical[0] + 4*c.Chemical[1])
	}
	t.Structures = structures(w)
	t.ActiveRules = RuleIdentity{Version: "ecology-2", Hash: "builtin"}
	if w.RuleState != nil && w.RuleState.Active != nil {
		t.ActiveRules = RuleIdentity{w.RuleState.Active.Source.Version, w.RuleState.Active.Hash}
	}
}

func structures(w *world.World) Structures {
	r := Structures{Sizes: []SizeCount{}, Bindings: []Binding{}}
	parent := make(map[uint64]uint64, len(w.Particles))
	for id := range w.Particles {
		parent[id] = id
	}
	find := func(id uint64) uint64 {
		for parent[id] != id {
			parent[id] = parent[parent[id]]
			id = parent[id]
		}
		return id
	}
	type pair struct{ a, b string }
	links := map[pair]int{}
	for _, b := range w.Relations {
		x, y := find(b.A), find(b.B)
		if x != y {
			parent[max(x, y)] = min(x, y)
		}
		a, c := w.Particles[b.A].Genome, w.Particles[b.B].Genome
		if a > c {
			a, c = c, a
		}
		links[pair{a, c}]++
	}
	sizes := map[uint64]int{}
	genomes := map[uint64]map[string]bool{}
	for id, p := range w.Particles {
		root := find(id)
		sizes[root]++
		if genomes[root] == nil {
			genomes[root] = map[string]bool{}
		}
		if p.Genome != "" {
			genomes[root][p.Genome] = true
		}
	}
	hist := map[int]int{}
	for root, n := range sizes {
		r.Components++
		hist[n]++
		r.Largest = max(r.Largest, n)
		if n > 1 {
			r.LinkedComponents++
			r.LinkedParticles += n
		}
		if len(genomes[root]) > 1 {
			r.MixedGenomeComponents++
		}
	}
	for size, count := range hist {
		r.Sizes = append(r.Sizes, SizeCount{size, count})
	}
	sort.Slice(r.Sizes, func(i, j int) bool { return r.Sizes[i].Size < r.Sizes[j].Size })
	for k, count := range links {
		r.Bindings = append(r.Bindings, Binding{k.a, k.b, count})
	}
	sort.Slice(r.Bindings, func(i, j int) bool {
		a, b := r.Bindings[i], r.Bindings[j]
		if a.A != b.A {
			return a.A < b.A
		}
		return a.B < b.B
	})
	return r
}

func sortEdges(e []Edge) {
	sort.Slice(e, func(i, j int) bool {
		a, b := e[i], e[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		return a.Target < b.Target
	})
}
