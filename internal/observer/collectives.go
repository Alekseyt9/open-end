package observer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"open-end/internal/rules"
	"open-end/internal/world"
	"slices"
	"sort"
	"strconv"
)

const collectiveLimit = 1024

type GroupView struct {
	Members []uint64  `json:"members"`
	Genomes []Lineage `json:"genomes"`
	Coded   int       `json:"coded"`
	Since   uint64    `json:"since_tick"`
}
type LineageActivity struct {
	Copies   uint64 `json:"copies"`
	Acquired int64  `json:"acquired_energy"`
	Shared   int64  `json:"bonded_transfer_energy"`
}
type CollectiveCohort struct {
	ID               int                        `json:"id"`
	Tagged           uint64                     `json:"tagged_tick"`
	Founders         []uint64                   `json:"founders"`
	FounderGenomes   []Lineage                  `json:"founder_genomes"`
	FoundersAlive    int                        `json:"founders_alive"`
	DescendantsAlive int                        `json:"descendants_alive"`
	Activity         map[string]LineageActivity `json:"tagged_lineage_activity"`
}
type DaughterCandidate struct {
	Cohort       int      `json:"cohort"`
	Members      []uint64 `json:"members"`
	Contributors int      `json:"founder_lineages"`
	AllFounders  bool     `json:"all_founder_lineages"`
	Since        uint64   `json:"since_tick"`
	Qualified    uint64   `json:"qualified_tick"`
	Last         uint64   `json:"last_observed_tick"`
	Copies       uint64   `json:"produced_copies"`
}
type CollectiveFrame struct {
	Version          int                 `json:"version"`
	ReferenceHash    string              `json:"reference_world_sha256"`
	Start            uint64              `json:"session_start_tick"`
	Tick             uint64              `json:"tick"`
	MinAge           uint64              `json:"minimum_age"`
	Limit            int                 `json:"tracking_limit"`
	Reference        []GroupView         `json:"reference_groups"`
	FoundersAlive    int                 `json:"reference_members_alive"`
	Intact           int                 `json:"reference_groups_together"`
	GroupTicks       uint64              `json:"group_ticks"`
	StableGroupTicks uint64              `json:"stable_group_ticks"`
	LinkedCopies     uint64              `json:"copies_by_linked_particles"`
	Shared           int64               `json:"bonded_transfer_energy"`
	MatureEvents     uint64              `json:"mature_membership_episodes"`
	Skipped          uint64              `json:"skipped_cohorts_or_candidates"`
	Cohorts          []CollectiveCohort  `json:"cohorts"`
	Candidates       []DaughterCandidate `json:"daughter_candidates"`
	Productive       int                 `json:"productive_daughter_candidates"`
	Current          []GroupView         `json:"current_groups"`
}
type founderTag struct {
	cohort  int
	founder uint64
}
type membershipAge struct {
	since             uint64
	mature, evaluated bool
}
type pendingDaughter struct {
	since, copies uint64
	index         int
	counted       bool
}
type collectiveTracker struct {
	w          *world.World
	state      CollectiveFrame
	tags       map[uint64]founderTag
	ages       map[string]*membershipAge
	pending    map[string]*pendingDaughter
	credited   map[string]int
	memberKeys map[uint64]string
}

// EnableCollectives adds read-only per-tick observation. A reference supplies
// the same initial groups to every intervention arm, including bond removal.
func (t *Tracker) EnableCollectives(w, reference *world.World, minAge uint64) error {
	if t.collectives != nil || w == nil || reference == nil || minAge == 0 || w.Tick != t.start || w.Tick != reference.Tick {
		return fmt.Errorf("invalid collective tracker initialization")
	}
	if err := reference.Validate(); err != nil {
		return err
	}
	b, err := json.Marshal(reference)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(b)
	c := &collectiveTracker{w: w, tags: map[uint64]founderTag{}, ages: map[string]*membershipAge{}, pending: map[string]*pendingDaughter{}, credited: map[string]int{}, memberKeys: map[uint64]string{}}
	c.state = CollectiveFrame{Version: 1, ReferenceHash: hex.EncodeToString(hash[:]), Start: w.Tick, Tick: w.Tick, MinAge: minAge, Limit: collectiveLimit, Reference: []GroupView{}, Cohorts: []CollectiveCohort{}, Candidates: []DaughterCandidate{}, Current: []GroupView{}}
	for _, g := range linkedGroups(reference) {
		if g.Coded == len(g.Members) {
			c.state.Reference = append(c.state.Reference, g)
		}
	}
	c.observe(false)
	t.collectives = c
	return nil
}
func (t *Tracker) Acquired(id uint64, genome string, energy int) {
	if c := t.collectives; c != nil {
		if tag, ok := c.tags[id]; ok {
			a := c.state.Cohorts[tag.cohort-1].Activity[genome]
			a.Acquired += int64(energy)
			c.state.Cohorts[tag.cohort-1].Activity[genome] = a
		}
	}
}
func groupKey(ids []uint64) string {
	b := []byte{}
	for _, id := range ids {
		b = strconv.AppendUint(b, id, 10)
		b = append(b, ',')
	}
	return string(b)
}

// Only graph-connected components of at least two particles are groups.
func linkedGroups(w *world.World) []GroupView {
	adjacent := map[uint64][]uint64{}
	for _, b := range w.Relations {
		adjacent[b.A] = append(adjacent[b.A], b.B)
		adjacent[b.B] = append(adjacent[b.B], b.A)
	}
	ids := make([]uint64, 0, len(adjacent))
	for id := range adjacent {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	visited := map[uint64]bool{}
	result := []GroupView{}
	for _, id := range ids {
		if visited[id] {
			continue
		}
		stack := []uint64{id}
		visited[id] = true
		members := []uint64{}
		genomes := map[string]int{}
		coded := 0
		for len(stack) > 0 {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			members = append(members, v)
			if p := w.Particles[v]; p != nil && len(p.Code) > 0 {
				coded++
				genomes[p.Genome]++
			}
			for _, next := range adjacent[v] {
				if !visited[next] {
					visited[next] = true
					stack = append(stack, next)
				}
			}
		}
		slices.Sort(members)
		g := GroupView{Members: members, Genomes: []Lineage{}, Coded: coded, Since: w.Tick}
		for hash, count := range genomes {
			g.Genomes = append(g.Genomes, Lineage{hash, count})
		}
		sort.Slice(g.Genomes, func(i, j int) bool { return g.Genomes[i].Hash < g.Genomes[j].Hash })
		result = append(result, g)
	}
	return result
}
func (c *collectiveTracker) interaction(e rules.Interaction) {
	if e.Kind == "transfer" && e.DirectBond {
		c.state.Shared += e.Energy
	}
	if e.Kind == "copy" {
		if c.w.Linked(e.SourceID) {
			c.state.LinkedCopies++
		}
		if tag, ok := c.tags[e.SourceID]; ok {
			c.tags[e.TargetID] = tag
		} else {
			delete(c.tags, e.TargetID)
		}
		// Exact membership is rechecked at the event, not inferred from a sample.
		if key := c.memberKeys[e.SourceID]; key != "" {
			if p := c.pending[key]; p != nil {
				for _, g := range linkedGroups(c.w) {
					if groupKey(g.Members) == key {
						p.copies++
						if p.counted {
							r := &c.state.Candidates[p.index]
							if r.Copies == 0 {
								c.state.Productive++
							}
							r.Copies++
						}
						break
					}
				}
			}
		}
	}
	if tag, ok := c.tags[e.SourceID]; ok {
		cohort := &c.state.Cohorts[tag.cohort-1]
		a := cohort.Activity[e.Source]
		if e.Kind == "copy" {
			a.Copies++
		}
		if e.Kind == "transfer" && e.DirectBond {
			a.Shared += e.Energy
		}
		cohort.Activity[e.Source] = a
	}
}
func (c *collectiveTracker) step() { c.observe(true) }
func (c *collectiveTracker) observe(advance bool) {
	groups := linkedGroups(c.w)
	keys := map[string]bool{}
	c.memberKeys = map[uint64]string{}
	for i := range groups {
		g := &groups[i]
		key := groupKey(g.Members)
		keys[key] = true
		age := c.ages[key]
		if age == nil {
			age = &membershipAge{since: c.w.Tick}
			c.ages[key] = age
		}
		g.Since = age.since
		for _, id := range g.Members {
			c.memberKeys[id] = key
		}
		if advance {
			c.state.GroupTicks++
		}
		if c.w.Tick-age.since < c.state.MinAge {
			continue
		}
		if advance {
			c.state.StableGroupTicks++
		}
		if !age.mature {
			c.state.MatureEvents++
			age.mature = true
		}
		if age.evaluated || g.Coded != len(g.Members) {
			continue
		}
		age.evaluated = true
		tagged := false
		for _, id := range g.Members {
			_, ok := c.tags[id]
			tagged = tagged || ok
		}
		if tagged {
			continue
		}
		if len(c.state.Cohorts) >= collectiveLimit {
			c.state.Skipped++
			continue
		}
		cohort := CollectiveCohort{ID: len(c.state.Cohorts) + 1, Tagged: c.w.Tick, Founders: slices.Clone(g.Members), FounderGenomes: slices.Clone(g.Genomes), Activity: map[string]LineageActivity{}}
		c.state.Cohorts = append(c.state.Cohorts, cohort)
		for _, id := range g.Members {
			c.tags[id] = founderTag{cohort.ID, id}
		}
	}
	for key := range c.ages {
		if !keys[key] {
			delete(c.ages, key)
		}
	}
	c.state.FoundersAlive = 0
	c.state.Intact = 0
	for _, g := range c.state.Reference {
		together := true
		key := c.memberKeys[g.Members[0]]
		for _, id := range g.Members {
			if c.w.Particles[id] != nil {
				c.state.FoundersAlive++
			}
			together = together && key != "" && c.memberKeys[id] == key
		}
		if together {
			c.state.Intact++
		}
	}
	for i := range c.state.Cohorts {
		cohort := &c.state.Cohorts[i]
		cohort.FoundersAlive = 0
		cohort.DescendantsAlive = 0
		for _, id := range cohort.Founders {
			if c.w.Particles[id] != nil {
				cohort.FoundersAlive++
			}
		}
	}
	for id, tag := range c.tags {
		if c.w.Particles[id] == nil {
			delete(c.tags, id)
			continue
		}
		cohort := &c.state.Cohorts[tag.cohort-1]
		if !slices.Contains(cohort.Founders, id) {
			cohort.DescendantsAlive++
		}
	}
	eligible := map[string]bool{}
	for _, g := range groups {
		if g.Coded != len(g.Members) {
			continue
		}
		tag, ok := c.tags[g.Members[0]]
		if !ok {
			continue
		}
		cohort := &c.state.Cohorts[tag.cohort-1]
		contributors := map[uint64]bool{}
		valid := true
		for _, id := range g.Members {
			t, ok := c.tags[id]
			if !ok || t.cohort != cohort.ID || slices.Contains(cohort.Founders, id) {
				valid = false
				break
			}
			contributors[t.founder] = true
		}
		if !valid || len(contributors) < 2 {
			continue
		}
		// A surviving parental component must retain at least two original founders.
		parents := map[string]int{}
		parentPresent := false
		for _, id := range cohort.Founders {
			if key := c.memberKeys[id]; key != "" {
				parents[key]++
				parentPresent = parentPresent || parents[key] >= 2
			}
		}
		if !parentPresent {
			continue
		}
		key := groupKey(g.Members)
		eligible[key] = true
		p := c.pending[key]
		if p == nil {
			p = &pendingDaughter{since: c.w.Tick}
			c.pending[key] = p
			if index, ok := c.credited[key]; ok {
				p.index = index
				p.counted = true
			}
		}
		if p.counted {
			c.state.Candidates[p.index].Last = c.w.Tick
			continue
		}
		if c.w.Tick-p.since < c.state.MinAge {
			continue
		}
		if len(c.state.Candidates) >= collectiveLimit {
			if p.index == 0 {
				c.state.Skipped++
				p.index = -1
			}
			continue
		}
		p.index = len(c.state.Candidates)
		p.counted = true
		c.credited[key] = p.index
		c.state.Candidates = append(c.state.Candidates, DaughterCandidate{Cohort: cohort.ID, Members: slices.Clone(g.Members), Contributors: len(contributors), AllFounders: len(contributors) == len(cohort.Founders), Since: p.since, Qualified: c.w.Tick, Last: c.w.Tick, Copies: p.copies})
		if p.copies > 0 {
			c.state.Productive++
		}
	}
	for key := range c.pending {
		if !eligible[key] {
			delete(c.pending, key)
		}
	}
	c.state.Tick = c.w.Tick
	c.state.Current = groups
}
func (c *collectiveTracker) frame() *CollectiveFrame {
	r := c.state
	r.Reference = cloneGroups(r.Reference)
	r.Current = cloneGroups(r.Current)
	r.Cohorts = slices.Clone(r.Cohorts)
	for i := range r.Cohorts {
		v := &r.Cohorts[i]
		v.Founders = slices.Clone(v.Founders)
		v.FounderGenomes = slices.Clone(v.FounderGenomes)
		v.Activity = maps.Clone(v.Activity)
	}
	r.Candidates = slices.Clone(r.Candidates)
	for i := range r.Candidates {
		r.Candidates[i].Members = slices.Clone(r.Candidates[i].Members)
	}
	return &r
}
func cloneGroups(groups []GroupView) []GroupView {
	r := slices.Clone(groups)
	for i := range r {
		r[i].Members = slices.Clone(r[i].Members)
		r[i].Genomes = slices.Clone(r[i].Genomes)
	}
	return r
}
