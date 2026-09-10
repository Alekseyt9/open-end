// Package discovery proposes descriptive entity boundaries outside world physics.
package discovery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/rules"
	"open-end/internal/world"
	"slices"
	"sort"
)

type Config struct {
	Every  int    `json:"every"`
	MinAge uint64 `json:"minimum_age"`
	Limit  int    `json:"interval_record_limit"`
}

func DefaultConfig() Config { return Config{Every: 1000, MinAge: 100, Limit: 65536} }

type Flow struct {
	Internal      int64    `json:"internal_transfer_energy"`
	Incoming      int64    `json:"incoming_transfer_energy"`
	Outgoing      int64    `json:"outgoing_transfer_energy"`
	TakenInternal int64    `json:"internal_take_energy"`
	TakenIn       int64    `json:"incoming_take_energy"`
	TakenOut      int64    `json:"outgoing_take_energy"`
	Acquired      int64    `json:"acquired_field_or_reaction_energy"`
	Copies        uint64   `json:"copies_by_members"`
	Retention     *float64 `json:"transfer_retention"`
}
type Node struct {
	ID                string   `json:"id"`
	Members           []uint64 `json:"members"`
	Sources           []string `json:"boundary_sources"`
	Children          []string `json:"direct_subset_candidates"`
	Level             int      `json:"structural_depth"`
	BondAge           uint64   `json:"bond_membership_age"`
	PersistentBond    bool     `json:"persistent_bond_boundary"`
	FlowSamples       int      `json:"consecutive_reciprocal_flow_samples"`
	FlowSpan          uint64   `json:"reciprocal_flow_observed_span"`
	PersistentFlow    bool     `json:"persistent_reciprocal_flow"`
	Overlap           int      `json:"overlapping_noncontained_candidates"`
	QualifiedDaughter bool     `json:"previously_qualified_daughter_membership"`
	Current           *Flow    `json:"discovery_interval_activity"`
	Next              *Flow    `json:"next_interval_activity"`
	NextFrom          uint64   `json:"next_interval_from,omitempty"`
	NextTo            uint64   `json:"next_interval_to,omitempty"`
	NextComplete      bool     `json:"next_interval_complete"`
}
type Frame struct {
	From     uint64 `json:"from_tick"`
	Tick     uint64 `json:"tick"`
	Micro    int    `json:"coded_micro_entities"`
	Complete bool   `json:"activity_complete"`
	Dropped  uint64 `json:"dropped_activity_records"`
	Nodes    []Node `json:"candidates"`
}
type Report struct {
	Version           int      `json:"version"`
	Config            Config   `json:"observation"`
	Seed              uint64   `json:"seed"`
	Treatment         string   `json:"treatment"`
	SourceHash        string   `json:"source_sha256"`
	InitialHash       string   `json:"initial_sha256"`
	FinalHash         string   `json:"final_sha256"`
	Frames            []Frame  `json:"frames"`
	Daughters         int      `json:"daughter_candidates"`
	Productive        int      `json:"productive_daughter_candidates"`
	CollectiveSkipped uint64   `json:"skipped_collective_records"`
	Limits            []string `json:"interpretation_limits"`
}
type pair struct{ a, b uint64 }
type edge struct{ transfer, taken int64 }
type activity struct {
	energy int64
	copies uint64
}
type episode struct {
	since   uint64
	samples int
}
type recorder struct {
	*observer.Tracker
	w            *world.World
	config       Config
	roots        map[uint64]uint64
	edges        map[pair]edge
	actors       map[uint64]activity
	flowEpisodes map[string]episode
	dropped      uint64
}

func newRecorder(w, reference *world.World, c Config) (*recorder, error) {
	if c.Every < 1 || c.MinAge == 0 || c.Limit < 1 || c.Limit > 1000000 {
		return nil, fmt.Errorf("invalid discovery observation configuration")
	}
	r := &recorder{Tracker: observer.NewTracker(w), w: w, config: c, roots: map[uint64]uint64{}, flowEpisodes: map[string]episode{}}
	if err := r.EnableCollectives(w, reference, c.MinAge); err != nil {
		return nil, err
	}
	for id, p := range w.Particles {
		if len(p.Code) > 0 {
			r.roots[id] = id
		}
	}
	r.reset()
	return r, nil
}
func (r *recorder) reset() {
	r.edges = map[pair]edge{}
	r.actors = map[uint64]activity{}
	r.dropped = 0
}
func (r *recorder) actor(id uint64, energy int64, copies uint64) {
	a, ok := r.actors[id]
	if !ok && len(r.actors) >= r.config.Limit {
		r.dropped++
		return
	}
	a.energy += energy
	a.copies += copies
	r.actors[id] = a
}
func (r *recorder) Acquired(id uint64, genome string, energy int) {
	r.Tracker.Acquired(id, genome, energy)
	r.actor(id, int64(energy), 0)
}
func (r *recorder) Death(d rules.Death) { r.Tracker.Death(d); delete(r.roots, d.ID) }
func (r *recorder) Interaction(e rules.Interaction) {
	r.Tracker.Interaction(e)
	if e.Kind == "copy" {
		if root, ok := r.roots[e.SourceID]; ok {
			r.roots[e.TargetID] = root
		} else {
			r.roots[e.TargetID] = e.TargetID
		}
		r.actor(e.SourceID, 0, 1)
	}
	if (e.Kind != "transfer" && e.Kind != "take") || e.Energy <= 0 {
		return
	}
	p := pair{e.SourceID, e.TargetID}
	v, ok := r.edges[p]
	if !ok && len(r.edges) >= r.config.Limit {
		r.dropped++
		return
	}
	if e.Kind == "transfer" {
		v.transfer += e.Energy
	} else {
		v.taken += e.Energy
	}
	r.edges[p] = v
}
func nodeID(ids []uint64) string {
	b, _ := json.Marshal(ids)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func (r *recorder) measure(ids []uint64) *Flow {
	if r.dropped > 0 {
		return nil
	}
	set := map[uint64]bool{}
	v := &Flow{}
	for _, id := range ids {
		set[id] = true
		a := r.actors[id]
		v.Acquired += a.energy
		v.Copies += a.copies
	}
	for p, e := range r.edges {
		a, b := set[p.a], set[p.b]
		switch {
		case a && b:
			v.Internal += e.transfer
			v.TakenInternal += e.taken
		case a:
			v.Outgoing += e.transfer
			v.TakenOut += e.taken
		case b:
			v.Incoming += e.transfer
			v.TakenIn += e.taken
		}
	}
	if total := v.Internal + v.Incoming + v.Outgoing; total > 0 {
		x := float64(v.Internal) / float64(total)
		v.Retention = &x
	}
	return v
}
func (r *recorder) sample(from uint64, previous *Frame) (Frame, *observer.CollectiveFrame) {
	f := Frame{From: from, Tick: r.w.Tick, Complete: r.dropped == 0, Dropped: r.dropped, Nodes: []Node{}}
	if previous != nil {
		for i := range previous.Nodes {
			n := &previous.Nodes[i]
			n.Next = r.measure(n.Members)
			n.NextFrom = from
			n.NextTo = r.w.Tick
			n.NextComplete = f.Complete
		}
	}
	g := r.Tracker.Frame(r.w).Telemetry.Collectives
	index := map[string]int{}
	add := func(ids []uint64, source string) *Node {
		ids = slices.Clone(ids)
		slices.Sort(ids)
		key := nodeID(ids)
		if i, ok := index[key]; ok {
			f.Nodes[i].Sources = append(f.Nodes[i].Sources, source)
			return &f.Nodes[i]
		}
		index[key] = len(f.Nodes)
		f.Nodes = append(f.Nodes, Node{ID: key, Members: ids, Sources: []string{source}, Children: []string{}, Level: 1})
		return &f.Nodes[len(f.Nodes)-1]
	}
	for _, b := range g.Current {
		if b.Coded != len(b.Members) {
			continue
		}
		n := add(b.Members, "bonds")
		n.BondAge = r.w.Tick - b.Since
		n.PersistentBond = n.BondAge >= r.config.MinAge
	}
	families := map[uint64][]uint64{}
	for id, p := range r.w.Particles {
		if len(p.Code) > 0 {
			f.Micro++
			families[r.roots[id]] = append(families[r.roots[id]], id)
		}
	}
	rootIDs := []uint64{}
	for root := range families {
		rootIDs = append(rootIDs, root)
	}
	slices.Sort(rootIDs)
	for _, id := range rootIDs {
		if len(families[id]) > 1 {
			add(families[id], "copy_ancestry")
		}
	}
	current := map[string]episode{}
	if f.Complete {
		for _, ids := range reciprocal(r.w, r.edges) {
			n := add(ids, "reciprocal_transfer")
			e, ok := r.flowEpisodes[n.ID]
			if !ok {
				e.since = r.w.Tick
			}
			e.samples++
			current[n.ID] = e
			n.FlowSamples = e.samples
			n.FlowSpan = r.w.Tick - e.since
			n.PersistentFlow = e.samples >= 2 && n.FlowSpan >= r.config.MinAge
		}
	}
	r.flowEpisodes = current
	for i := range f.Nodes {
		n := &f.Nodes[i]
		n.Current = r.measure(n.Members)
		for _, d := range g.Candidates {
			if slices.Equal(n.Members, d.Members) {
				n.QualifiedDaughter = true
			}
		}
		slices.Sort(n.Sources)
	}
	sort.Slice(f.Nodes, func(i, j int) bool {
		a, b := f.Nodes[i], f.Nodes[j]
		if len(a.Members) != len(b.Members) {
			return len(a.Members) < len(b.Members)
		}
		return a.ID < b.ID
	})
	hierarchy(f.Nodes)
	r.reset()
	return f, g
}

// Replay reproduces an assay treatment, collecting new evidence without changing physics.
func Replay(source *world.World, treatment string, ticks int, c Config) (Report, error) {
	out := Report{Version: 1, Config: c, Treatment: treatment, Frames: []Frame{}, Limits: []string{
		"Candidates and subset depth are descriptive; no higher-level individuality is established.",
		"Information closure and causal coordination are unmeasured. Transfer retention excludes field inflow and is not energetic autonomy.",
		"Flow and ancestry boundaries are sampled; only bond membership ages use every tick boundary.",
		"Discovery activity uses endpoint member IDs retrospectively; next-interval activity freezes those IDs prospectively, including members that later die.",
		"Flow discovery selects on the same interval's transfers; use the separate next interval to assess persistence.",
		"COPY ancestry starts with distinct coded particles at replay start; historical common ancestry is not reconstructed.",
	}}
	if ticks < 1 {
		return out, fmt.Errorf("positive replay duration required")
	}
	w, err := experiment.CollectiveStart(source, treatment)
	if err != nil {
		return out, err
	}
	r, err := newRecorder(w, source, c)
	if err != nil {
		return out, err
	}
	out.SourceHash = kernel.Hash(source)
	out.InitialHash = kernel.Hash(w)
	out.Seed = w.Config.Seed
	f, _ := r.sample(w.Tick, nil)
	out.Frames = append(out.Frames, f)
	for i := 1; i <= ticks; i++ {
		kernel.StepObserved(w, r)
		if i%c.Every == 0 || i == ticks {
			last := &out.Frames[len(out.Frames)-1]
			f, g := r.sample(last.Tick, last)
			out.Frames = append(out.Frames, f)
			out.Daughters = len(g.Candidates)
			out.Productive = g.Productive
			out.CollectiveSkipped = g.Skipped
		}
	}
	if err := w.Validate(); err != nil {
		return out, err
	}
	out.FinalHash = kernel.Hash(w)
	return out, nil
}
