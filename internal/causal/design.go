// Package causal separates predictive descriptions from physical interventions.
package causal

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"open-end/internal/discovery"
	"open-end/internal/kernel"
	"open-end/internal/rules"
	"open-end/internal/world"
	"slices"
	"sort"
)

type Config struct {
	Horizon      int     `json:"horizon"`
	Blocks       int     `json:"forecast_blocks"`
	Permutations int     `json:"random_partitions"`
	Ridge        float64 `json:"ridge"`
}

func DefaultConfig() Config { return Config{1000, 5, 3, 0.01} }
func (c Config) Validate() error {
	if c.Horizon < 1 || c.Blocks < 1 || c.Blocks > 100 || c.Permutations < 1 || c.Permutations > 16 || c.Ridge <= 0 || math.IsNaN(c.Ridge) || math.IsInf(c.Ridge, 0) {
		return fmt.Errorf("invalid causal analysis configuration")
	}
	return nil
}

type Group struct {
	ID        string   `json:"id"`
	Partition string   `json:"partition"`
	Members   []uint64 `json:"members"`
}
type Member struct {
	ID    uint64    `json:"id"`
	X     []float64 `json:"features"`
	Alive float64   `json:"alive_next"`
}
type Sample struct {
	Seed      uint64    `json:"seed"`
	Group     string    `json:"group"`
	Partition string    `json:"partition"`
	From      uint64    `json:"from_tick"`
	To        uint64    `json:"to_tick"`
	X         []float64 `json:"macro_features"`
	Members   []Member  `json:"members_at_prediction"`
	Y         float64   `json:"next_survival_fraction"`
}

var MicroFeatures = []string{"energy/capacity", "age/(age+1000)", "code_length/max_code", "field_energy/capacity", "matter/(matter+16)", "X/(X+16)", "Y/(Y+16)", "Z/(Z+16)", "terrain/8", "signal/64", "bond_degree/4", "occupied_neighbors/4"}
var MacroFeatures = append(append([]string{}, MicroFeatures...), "size/(size+8)", "energy_variance", "minimum_energy", "genome_dominance", "internal_bonds/(2*size)", "external_bonds/(4*size)", "mean_pair_toroidal_distance")

func clone(w *world.World) (*world.World, error) {
	var b bytes.Buffer
	if err := kernel.Save(&b, w); err != nil {
		return nil, err
	}
	return kernel.Load(&b)
}
func identity(ids []uint64) string {
	b, _ := json.Marshal(ids)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
func rank(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }

// Select uses only evidence available at the source tick. Targets are disjoint,
// complete physical bond components, rather than arbitrary contained subsets.
func Select(w *world.World, report discovery.Report, c Config) ([]Group, error) {
	if w == nil {
		return nil, fmt.Errorf("missing source world")
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	if report.Version != 1 || report.Treatment != "intact" || w.Config.CollectiveAblation != "" || report.FinalHash != kernel.Hash(w) || len(report.Frames) == 0 || report.Config.MinAge == 0 {
		return nil, fmt.Errorf("unmatched discovery/source")
	}
	f := report.Frames[len(report.Frames)-1]
	if f.Tick != w.Tick {
		return nil, fmt.Errorf("discovery tick mismatch")
	}
	coded := []uint64{}
	for id, p := range w.Particles {
		if len(p.Code) > 0 {
			coded = append(coded, id)
		}
	}
	slices.Sort(coded)
	if !slices.Equal(coded, f.MicroIDs) {
		return nil, fmt.Errorf("discovery micro membership mismatch")
	}
	groups := []Group{}
	used := map[uint64]bool{}
	for _, n := range f.Nodes {
		if !n.PersistentBond {
			continue
		}
		if !slices.Contains(n.Sources, "bonds") || n.BondAge < report.Config.MinAge || len(n.Members) < 2 || !slices.IsSorted(n.Members) || n.ID != identity(n.Members) {
			return nil, fmt.Errorf("invalid selected boundary")
		}
		set := map[uint64]bool{}
		for _, id := range n.Members {
			p := w.Particles[id]
			if p == nil || len(p.Code) == 0 || used[id] || set[id] {
				return nil, fmt.Errorf("invalid or overlapping selected member")
			}
			set[id] = true
			used[id] = true
		}
		adj := map[uint64][]uint64{}
		for _, b := range w.Relations {
			if set[b.A] != set[b.B] {
				return nil, fmt.Errorf("boundary is not a complete bond component")
			}
			if set[b.A] {
				adj[b.A] = append(adj[b.A], b.B)
				adj[b.B] = append(adj[b.B], b.A)
			}
		}
		seen := map[uint64]bool{n.Members[0]: true}
		queue := []uint64{n.Members[0]}
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			for _, u := range adj[v] {
				if !seen[u] {
					seen[u] = true
					queue = append(queue, u)
				}
			}
		}
		if len(seen) != len(set) {
			return nil, fmt.Errorf("selected component disconnected")
		}
		groups = append(groups, Group{n.ID, "bonds", slices.Clone(n.Members)})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].ID < groups[j].ID })
	pool := []uint64{}
	for id := range used {
		pool = append(pool, id)
	}
	slices.Sort(pool)
	actual := slices.Clone(groups)
	for repetition := 1; repetition <= c.Permutations; repetition++ {
		partition := fmt.Sprintf("random-%d", repetition)
		shuffled := slices.Clone(pool)
		keys := map[uint64]string{}
		for _, id := range shuffled {
			keys[id] = rank(fmt.Sprintf("%s/%s/%d", report.FinalHash, partition, id))
		}
		sort.Slice(shuffled, func(i, j int) bool {
			a, b := shuffled[i], shuffled[j]
			if keys[a] != keys[b] {
				return keys[a] < keys[b]
			}
			return a < b
		})
		from := 0
		for _, g := range actual {
			members := slices.Clone(shuffled[from : from+len(g.Members)])
			slices.Sort(members)
			from += len(g.Members)
			groups = append(groups, Group{g.ID, partition, members})
		}
	}
	return groups, nil
}

func features(w *world.World, id uint64) []float64 {
	p := w.Particles[id]
	cell := w.Cells[p.Position]
	degree, occupied := 0, 0
	for _, b := range w.Relations {
		if b.A == id || b.B == id {
			degree++
		}
	}
	for d := 0; d < 4; d++ {
		if w.Cells[w.Neighbor(p.Position, d)].Occupant != 0 {
			occupied++
		}
	}
	sat := func(x int) float64 { return float64(x) / float64(x+16) }
	age := float64(w.Tick - p.Created)
	return []float64{float64(p.Energy) / float64(w.Config.EnergyCapacity), age / (age + 1000), float64(len(p.Code)) / float64(w.Config.MaxCode), float64(cell.Energy) / float64(w.Config.CellCapacity), sat(cell.Matter), sat(cell.Chemical[0]), sat(cell.Chemical[1]), sat(cell.Chemical[2]), float64(cell.Terrain) / 8, float64(cell.Signal) / 64, float64(degree) / 4, float64(occupied) / 4}
}
func observe(w *world.World, g Group) Sample {
	s := Sample{Seed: w.Config.Seed, Group: g.ID, Partition: g.Partition, From: w.Tick, Members: []Member{}, X: make([]float64, len(MacroFeatures))}
	set := map[uint64]bool{}
	genomes := map[string]int{}
	minEnergy := 1.0
	for _, id := range g.Members {
		if p := w.Particles[id]; p != nil && len(p.Code) > 0 {
			x := features(w, id)
			s.Members = append(s.Members, Member{ID: id, X: x})
			set[id] = true
			genomes[p.Genome]++
			minEnergy = min(minEnergy, x[0])
			for i, v := range x {
				s.X[i] += v
			}
		}
	}
	if len(s.Members) == 0 {
		return s
	}
	size := float64(len(s.Members))
	for i := range MicroFeatures {
		s.X[i] /= size
	}
	variance := 0.0
	for _, m := range s.Members {
		variance += (m.X[0] - s.X[0]) * (m.X[0] - s.X[0]) / size
	}
	dominant := 0
	for _, n := range genomes {
		dominant = max(dominant, n)
	}
	internal, external := 0, 0
	for _, b := range w.Relations {
		if set[b.A] && set[b.B] {
			internal++
		} else if set[b.A] || set[b.B] {
			external++
		}
	}
	distance := 0.0
	pairs := 0
	for i, a := range s.Members {
		pa := w.Particles[a.ID].Position
		for _, b := range s.Members[i+1:] {
			pb := w.Particles[b.ID].Position
			dx := int(math.Abs(float64(pa%w.Config.Width - pb%w.Config.Width)))
			dy := int(math.Abs(float64(pa/w.Config.Width - pb/w.Config.Width)))
			distance += float64(min(dx, w.Config.Width-dx)+min(dy, w.Config.Height-dy)) / float64(w.Config.Width/2+w.Config.Height/2)
			pairs++
		}
	}
	if pairs > 0 {
		distance /= float64(pairs)
	}
	copy(s.X[len(MicroFeatures):], []float64{size / (size + 8), variance, minEnergy, float64(dominant) / size, float64(internal) / (2 * size), float64(external) / (4 * size), distance})
	return s
}

type Outcome struct {
	Members          int    `json:"members"`
	Alive            int    `json:"alive"`
	Copies           uint64 `json:"copies_by_original_members"`
	InternalTransfer int64  `json:"internal_transfer_energy"`
	InternalBonds    int    `json:"internal_bonds_at_end"`
}
type counter struct {
	members  map[uint64]bool
	copies   uint64
	transfer int64
}

func (c *counter) Interaction(e rules.Interaction) {
	if e.Kind == "copy" && c.members[e.SourceID] {
		c.copies++
	}
	if e.Kind == "transfer" && c.members[e.SourceID] && c.members[e.TargetID] {
		c.transfer += e.Energy
	}
}
func (*counter) Death(rules.Death)    {}
func (*counter) TickCompleted(uint64) {}
func outcome(w *world.World, g Group, c *counter) Outcome {
	v := Outcome{Members: len(g.Members), Copies: c.copies, InternalTransfer: c.transfer}
	for _, id := range g.Members {
		if w.Particles[id] != nil {
			v.Alive++
		}
	}
	for _, b := range w.Relations {
		if c.members[b.A] && c.members[b.B] {
			v.InternalBonds++
		}
	}
	return v
}
func newCounter(g Group) *counter {
	c := &counter{members: map[uint64]bool{}}
	for _, id := range g.Members {
		c.members[id] = true
	}
	return c
}

type counters []*counter

func (cs counters) Interaction(e rules.Interaction) {
	for _, c := range cs {
		c.Interaction(e)
	}
}
func (counters) Death(rules.Death)    {}
func (counters) TickCompleted(uint64) {}

type Baseline struct {
	Seed       uint64             `json:"seed"`
	SourceHash string             `json:"source_sha256"`
	Groups     []Group            `json:"groups"`
	Samples    []Sample           `json:"forecast_samples"`
	FirstHash  string             `json:"first_horizon_hash"`
	FinalHash  string             `json:"final_hash"`
	Control    map[string]Outcome `json:"first_horizon_outcomes"`
}

func Forecast(source *world.World, groups []Group, c Config) (Baseline, error) {
	if source == nil {
		return Baseline{}, fmt.Errorf("missing source world")
	}
	r := Baseline{Seed: source.Config.Seed, SourceHash: kernel.Hash(source), Groups: groups, Samples: []Sample{}, Control: map[string]Outcome{}}
	if err := c.Validate(); err != nil {
		return r, err
	}
	w, err := clone(source)
	if err != nil {
		return r, err
	}
	cs := counters{}
	actual := []Group{}
	for _, g := range groups {
		if g.Partition == "bonds" {
			actual = append(actual, g)
			cs = append(cs, newCounter(g))
		}
	}
	for block := 0; block < c.Blocks; block++ {
		pending := []Sample{}
		for _, g := range groups {
			s := observe(w, g)
			if len(s.Members) >= 2 {
				pending = append(pending, s)
			}
		}
		for tick := 0; tick < c.Horizon; tick++ {
			if block == 0 {
				kernel.StepObserved(w, cs)
			} else {
				kernel.Step(w)
			}
		}
		for _, s := range pending {
			s.To = w.Tick
			for i := range s.Members {
				if w.Particles[s.Members[i].ID] != nil {
					s.Members[i].Alive = 1
					s.Y++
				}
			}
			s.Y /= float64(len(s.Members))
			r.Samples = append(r.Samples, s)
		}
		if block == 0 {
			r.FirstHash = kernel.Hash(w)
			for i, g := range actual {
				r.Control[g.ID] = outcome(w, g, cs[i])
			}
		}
	}
	if err := w.Validate(); err != nil {
		return r, err
	}
	r.FinalHash = kernel.Hash(w)
	return r, nil
}

type Intervention struct {
	Seed        uint64           `json:"seed"`
	Group       string           `json:"group"`
	Mode        string           `json:"mode"`
	Available   bool             `json:"available"`
	Reason      string           `json:"unavailable_reason,omitempty"`
	SourceHash  string           `json:"source_sha256"`
	InitialHash string           `json:"initial_sha256,omitempty"`
	FinalHash   string           `json:"final_sha256,omitempty"`
	Removed     []world.Relation `json:"removed_bonds"`
	Outcome     Outcome          `json:"outcome"`
}

// Intervene is a one-time bond cut. Ordinary BIND can subsequently restore links.
func Intervene(source *world.World, g Group, mode string, horizon int) (Intervention, error) {
	if source == nil {
		return Intervention{}, fmt.Errorf("missing source world")
	}
	r := Intervention{Seed: source.Config.Seed, Group: g.ID, Mode: mode, SourceHash: kernel.Hash(source), Removed: []world.Relation{}}
	if horizon < 1 || (mode != "local-cut" && mode != "outside-cut") {
		return r, fmt.Errorf("invalid intervention")
	}
	member := newCounter(g)
	inside, outside := []world.Relation{}, []world.Relation{}
	for _, b := range source.Relations {
		if member.members[b.A] && member.members[b.B] {
			inside = append(inside, b)
		} else if !member.members[b.A] && !member.members[b.B] {
			outside = append(outside, b)
		}
	}
	if len(inside) == 0 {
		return r, fmt.Errorf("target has no internal bonds")
	}
	if mode == "local-cut" {
		r.Removed = inside
	} else {
		if len(outside) < len(inside) {
			r.Reason = "fewer outside bonds than target internal bonds"
			return r, nil
		}
		sort.Slice(outside, func(i, j int) bool {
			a, b := outside[i], outside[j]
			ka := rank(fmt.Sprintf("%s/%s/%d/%d", r.SourceHash, g.ID, a.A, a.B))
			kb := rank(fmt.Sprintf("%s/%s/%d/%d", r.SourceHash, g.ID, b.A, b.B))
			if ka != kb {
				return ka < kb
			}
			if a.A != b.A {
				return a.A < b.A
			}
			return a.B < b.B
		})
		r.Removed = slices.Clone(outside[:len(inside)])
	}
	sort.Slice(r.Removed, func(i, j int) bool {
		a, b := r.Removed[i], r.Removed[j]
		if a.A != b.A {
			return a.A < b.A
		}
		return a.B < b.B
	})
	w, err := clone(source)
	if err != nil {
		return r, err
	}
	for _, b := range r.Removed {
		delete(w.Relations, world.RelationKey(b.A, b.B))
	}
	if err := w.Validate(); err != nil {
		return r, err
	}
	r.Available = true
	r.InitialHash = kernel.Hash(w)
	for tick := 0; tick < horizon; tick++ {
		kernel.StepObserved(w, member)
	}
	if err := w.Validate(); err != nil {
		return r, err
	}
	r.FinalHash = kernel.Hash(w)
	r.Outcome = outcome(w, g, member)
	return r, nil
}
