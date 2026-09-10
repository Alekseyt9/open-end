package observer

import "slices"

// DescendantBranch exposes ancestry conditions without changing the original
// multi-founder candidate definition or any serialized collective frame.
type DescendantBranch struct {
	Cohort        int       `json:"cohort"`
	Members       []uint64  `json:"members"`
	Contributors  int       `json:"founder_lineages"`
	ParentPresent bool      `json:"parent_present"`
	FounderIDs    []uint64  `json:"founder_ids"`
	Genomes       []Lineage `json:"genomes"`
}

// DescendantBranches returns fully coded, founder-free connected groups whose
// members all descend from one observed cohort, including one-founder branches.
func (t *Tracker) DescendantBranches() []DescendantBranch {
	result := []DescendantBranch{}
	c := t.collectives
	if c == nil {
		return result
	}
	for _, g := range c.state.Current {
		if g.Coded != len(g.Members) {
			continue
		}
		tag, ok := c.tags[g.Members[0]]
		if !ok {
			continue
		}
		cohort := c.state.Cohorts[tag.cohort-1]
		contributors := map[uint64]bool{}
		valid := true
		for _, id := range g.Members {
			x, ok := c.tags[id]
			if !ok || x.cohort != tag.cohort || slices.Contains(cohort.Founders, id) {
				valid = false
				break
			}
			contributors[x.founder] = true
		}
		if !valid {
			continue
		}
		parents := map[string]int{}
		present := false
		for _, id := range cohort.Founders {
			if key := c.memberKeys[id]; key != "" {
				parents[key]++
				present = present || parents[key] >= 2
			}
		}
		founders := make([]uint64, 0, len(contributors))
		for id := range contributors {
			founders = append(founders, id)
		}
		slices.Sort(founders)
		result = append(result, DescendantBranch{Cohort: tag.cohort, Members: slices.Clone(g.Members), Contributors: len(contributors), ParentPresent: present, FounderIDs: founders, Genomes: slices.Clone(g.Genomes)})
	}
	return result
}
