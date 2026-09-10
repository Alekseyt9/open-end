package observer

import (
	"fmt"
	"reflect"
	"slices"
)

type CollectiveSummary struct {
	Ticks            uint64           `json:"ticks"`
	MeanGroups       float64          `json:"mean_linked_groups"`
	StableGroupTicks uint64           `json:"stable_group_ticks"`
	LinkedCopies     uint64           `json:"copies_by_linked_particles"`
	Shared           int64            `json:"bonded_transfer_energy"`
	NewCohorts       int              `json:"new_tracked_cohorts"`
	NewCandidates    int              `json:"new_daughter_candidates"`
	NewProductive    int              `json:"new_productive_daughter_candidates"`
	End              *CollectiveFrame `json:"end"`
}

func validateGroups(groups []GroupView, tick uint64) error {
	seen := map[uint64]bool{}
	for _, g := range groups {
		if len(g.Members) < 2 || g.Since > tick || g.Coded < 0 || g.Coded > len(g.Members) {
			return fmt.Errorf("invalid group")
		}
		prev := uint64(0)
		for _, id := range g.Members {
			if id <= prev || seen[id] {
				return fmt.Errorf("invalid or duplicate group member")
			}
			seen[id] = true
			prev = id
		}
		coded := 0
		hash := ""
		for _, lineage := range g.Genomes {
			if len(lineage.Hash) != 64 || lineage.Hash <= hash || lineage.Count < 1 {
				return fmt.Errorf("invalid group genomes")
			}
			hash = lineage.Hash
			coded += lineage.Count
		}
		if coded != g.Coded {
			return fmt.Errorf("group composition mismatch")
		}
	}
	return nil
}
func validateCollectiveFrame(c *CollectiveFrame, tick uint64) error {
	if c.Version != 1 || len(c.ReferenceHash) != 64 || c.Tick != tick || c.Tick < c.Start || c.MinAge == 0 || c.Limit != collectiveLimit || len(c.Cohorts) > c.Limit || len(c.Candidates) > c.Limit || c.Shared < 0 || c.StableGroupTicks > c.GroupTicks {
		return fmt.Errorf("invalid collective frame")
	}
	if err := validateGroups(c.Reference, c.Start); err != nil {
		return err
	}
	if err := validateGroups(c.Current, c.Tick); err != nil {
		return err
	}
	count := 0
	for _, g := range c.Reference {
		if g.Coded != len(g.Members) {
			return fmt.Errorf("uncoded reference group")
		}
		count += len(g.Members)
	}
	if c.FoundersAlive < 0 || c.FoundersAlive > count || c.Intact < 0 || c.Intact > len(c.Reference) {
		return fmt.Errorf("invalid reference survival")
	}
	founders := map[uint64]bool{}
	for i, g := range c.Cohorts {
		if g.ID != i+1 || g.Tagged < c.Start || g.Tagged > c.Tick || g.Tagged-c.Start < c.MinAge || g.FoundersAlive < 0 || g.FoundersAlive > len(g.Founders) || g.DescendantsAlive < 0 {
			return fmt.Errorf("invalid cohort")
		}
		if err := validateGroups([]GroupView{{Members: g.Founders, Genomes: g.FounderGenomes, Coded: len(g.Founders)}}, c.Tick); err != nil {
			return err
		}
		for _, id := range g.Founders {
			if founders[id] {
				return fmt.Errorf("founder reassigned")
			}
			founders[id] = true
		}
		for hash, a := range g.Activity {
			if len(hash) != 64 || a.Acquired < 0 || a.Shared < 0 {
				return fmt.Errorf("invalid lineage activity")
			}
		}
	}
	productive := 0
	credited := map[string]bool{}
	for _, d := range c.Candidates {
		if d.Cohort < 1 || d.Cohort > len(c.Cohorts) || len(d.Members) < 2 || d.Contributors < 2 || d.Contributors > len(d.Members) || d.Since < c.Start || d.Since > d.Qualified || d.Qualified-d.Since < c.MinAge || d.Qualified > d.Last || d.Last > c.Tick {
			return fmt.Errorf("invalid daughter candidate")
		}
		g := c.Cohorts[d.Cohort-1]
		if d.Contributors > len(g.Founders) || d.AllFounders != (d.Contributors == len(g.Founders)) {
			return fmt.Errorf("invalid founder coverage")
		}
		prev := uint64(0)
		for _, id := range d.Members {
			if id <= prev || slices.Contains(g.Founders, id) {
				return fmt.Errorf("daughter contains parent founder")
			}
			prev = id
		}
		key := groupKey(d.Members)
		if credited[key] {
			return fmt.Errorf("duplicate daughter candidate")
		}
		credited[key] = true
		if d.Copies > 0 {
			productive++
		}
	}
	if productive != c.Productive {
		return fmt.Errorf("productive candidate count mismatch")
	}
	return nil
}
func summarizeCollectives(frames []Metrics) (*CollectiveSummary, error) {
	get := func(m Metrics) *CollectiveFrame {
		if m.Telemetry == nil {
			return nil
		}
		return m.Telemetry.Collectives
	}
	first := get(frames[0])
	if first == nil {
		for _, m := range frames {
			if get(m) != nil {
				return nil, fmt.Errorf("collective observation starts mid-window")
			}
		}
		return nil, nil
	}
	previous := first
	for _, m := range frames {
		c := get(m)
		if c == nil {
			return nil, fmt.Errorf("missing collective frame")
		}
		if err := validateCollectiveFrame(c, m.Tick); err != nil {
			return nil, err
		}
		if c.ReferenceHash != first.ReferenceHash || c.Start != first.Start || c.MinAge != first.MinAge || !reflect.DeepEqual(c.Reference, first.Reference) {
			return nil, fmt.Errorf("collective session changed")
		}
		if c.GroupTicks < previous.GroupTicks || c.StableGroupTicks < previous.StableGroupTicks || c.LinkedCopies < previous.LinkedCopies || c.Shared < previous.Shared || c.MatureEvents < previous.MatureEvents || c.Skipped < previous.Skipped || len(c.Cohorts) < len(previous.Cohorts) || len(c.Candidates) < len(previous.Candidates) || c.Productive < previous.Productive {
			return nil, fmt.Errorf("collective history regressed")
		}
		for i, old := range previous.Cohorts {
			g := c.Cohorts[i]
			if g.Tagged != old.Tagged || !slices.Equal(g.Founders, old.Founders) || !reflect.DeepEqual(g.FounderGenomes, old.FounderGenomes) {
				return nil, fmt.Errorf("cohort founders changed")
			}
			for hash, a := range old.Activity {
				b, ok := g.Activity[hash]
				if !ok || b.Copies < a.Copies || b.Acquired < a.Acquired || b.Shared < a.Shared {
					return nil, fmt.Errorf("cohort activity regressed")
				}
			}
		}
		for i, old := range previous.Candidates {
			d := c.Candidates[i]
			if d.Cohort != old.Cohort || !slices.Equal(d.Members, old.Members) || d.Contributors != old.Contributors || d.Since != old.Since || d.Qualified != old.Qualified || d.Last < old.Last || d.Copies < old.Copies {
				return nil, fmt.Errorf("daughter history changed")
			}
		}
		previous = c
	}
	last := previous
	ticks := last.Tick - first.Tick
	// Own all nested state, just as ordinary tracker frames do.
	copy := (&collectiveTracker{state: *last}).frame()
	r := &CollectiveSummary{Ticks: ticks, StableGroupTicks: last.StableGroupTicks - first.StableGroupTicks, LinkedCopies: last.LinkedCopies - first.LinkedCopies, Shared: last.Shared - first.Shared, NewCohorts: len(last.Cohorts) - len(first.Cohorts), NewCandidates: len(last.Candidates) - len(first.Candidates), NewProductive: last.Productive - first.Productive, End: copy}
	if ticks > 0 {
		r.MeanGroups = float64(last.GroupTicks-first.GroupTicks) / float64(ticks)
	}
	return r, nil
}
