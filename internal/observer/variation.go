package observer

import (
	"fmt"
	"open-end/internal/evolution"
	"sort"
)

type VariationActivity struct {
	Genome        string               `json:"genome"`
	Policy        evolution.CopyPolicy `json:"policy"`
	Copies        uint64               `json:"copies"`
	Changed       uint64               `json:"changed"`
	Recombined    uint64               `json:"recombined"`
	Donors        map[string]uint64    `json:"donors,omitempty"`
	EndPopulation int                  `json:"end_population"`
	Persistent    bool                 `json:"persistent"`
}
type VariationWindow struct {
	Model                  string `json:"model"`
	CodePolicies           int    `json:"code_policies"`
	MemoryPolicies         int    `json:"memory_policies"`
	PersistentCodePolicies int    `json:"persistent_code_policies"`
	CodeCopies             uint64 `json:"code_copies"`
	ChangedCode            uint64 `json:"changed_code"`
	Recombined             uint64 `json:"recombined"`
	// Persistent requires >=5 individuals at every sample and >=1 successful
	// copy by the same genome/policy in every reporting interval.
	Activity []VariationActivity `json:"activity"`
}

func variationRecords(m Metrics) (map[string]evolution.CopyRecord, error) {
	records := map[string]evolution.CopyRecord{}
	var total uint64
	for _, r := range m.Variation {
		key := evolution.PolicyKey(r.Genome, r.Policy)
		if !r.Policy.Valid() || len(r.Genome) != 64 || r.Copies == 0 || r.Changed > r.Copies || r.Recombined > r.Copies || r.FirstTick > r.LastTick || r.LastTick > m.Tick {
			return nil, fmt.Errorf("invalid variation counters at tick %d", m.Tick)
		}
		if _, ok := records[key]; ok {
			return nil, fmt.Errorf("duplicate variation record")
		}
		var donors uint64
		for hash, n := range r.Donors {
			if len(hash) != 64 || n == 0 || n > r.Recombined-donors {
				return nil, fmt.Errorf("invalid variation donors")
			}
			donors += n
		}
		if donors != r.Recombined || donors > 0 && !r.Policy.Recombine {
			return nil, fmt.Errorf("variation donors do not reconcile")
		}
		if r.Policy.Kind == "code" {
			if r.Copies > m.Copies-total {
				return nil, fmt.Errorf("variation copy total exceeds accounting")
			}
			total += r.Copies
		}
		records[key] = r
	}
	if total != m.Copies {
		return nil, fmt.Errorf("variation copies do not reconcile at tick %d", m.Tick)
	}
	return records, nil
}

func summarizeVariation(frames []Metrics) (*VariationWindow, error) {
	model := frames[0].CopyModel
	for _, m := range frames {
		if m.CopyModel != model || model == "" && len(m.Variation) > 0 {
			return nil, fmt.Errorf("inconsistent copy model")
		}
	}
	if model == "" {
		return nil, nil
	}
	if model != "fixed" && model != "evolving" {
		return nil, fmt.Errorf("unknown copy model")
	}
	r := &VariationWindow{Model: model, Activity: []VariationActivity{}}
	first, err := variationRecords(frames[0])
	if err != nil {
		return nil, err
	}
	previous := first
	persistent := map[string]bool{}
	population := func(m Metrics) map[string]int {
		p := map[string]int{}
		for _, g := range m.ActiveGenomes {
			p[g.Hash] = g.Count
		}
		return p
	}
	pop := population(frames[0])
	for key, record := range first {
		persistent[key] = pop[record.Genome] >= 5
	}
	for _, m := range frames[1:] {
		current, err := variationRecords(m)
		if err != nil {
			return nil, err
		}
		pop = population(m)
		for key, old := range previous {
			rec, ok := current[key]
			if !ok || rec.Copies < old.Copies || rec.Changed < old.Changed || rec.Recombined < old.Recombined || rec.FirstTick != old.FirstTick || rec.LastTick < old.LastTick || rec.Changed-old.Changed > rec.Copies-old.Copies || rec.Recombined-old.Recombined > rec.Copies-old.Copies {
				return nil, fmt.Errorf("variation ledger regressed at tick %d", m.Tick)
			}
			for donor, n := range old.Donors {
				if rec.Donors[donor] < n {
					return nil, fmt.Errorf("donor ledger regressed")
				}
			}
			persistent[key] = persistent[key] && pop[rec.Genome] >= 5 && rec.Copies > old.Copies
		}
		previous = current
	}
	code, memory, stable := map[evolution.CopyPolicy]bool{}, map[evolution.CopyPolicy]bool{}, map[evolution.CopyPolicy]bool{}
	for key, rec := range previous {
		old := first[key]
		copies := rec.Copies - old.Copies
		if copies == 0 {
			continue
		}
		a := VariationActivity{Genome: rec.Genome, Policy: rec.Policy, Copies: copies, Changed: rec.Changed - old.Changed, Recombined: rec.Recombined - old.Recombined, EndPopulation: pop[rec.Genome], Persistent: persistent[key]}
		for donor, n := range rec.Donors {
			if n > old.Donors[donor] {
				if a.Donors == nil {
					a.Donors = map[string]uint64{}
				}
				a.Donors[donor] = n - old.Donors[donor]
			}
		}
		r.Activity = append(r.Activity, a)
		if rec.Policy.Kind == "code" {
			code[rec.Policy] = true
			r.CodeCopies += copies
			r.ChangedCode += a.Changed
			r.Recombined += a.Recombined
			if a.Persistent {
				stable[rec.Policy] = true
			}
		} else {
			memory[rec.Policy] = true
		}
	}
	r.CodePolicies = len(code)
	r.MemoryPolicies = len(memory)
	r.PersistentCodePolicies = len(stable)
	sort.Slice(r.Activity, func(i, j int) bool {
		a, b := r.Activity[i], r.Activity[j]
		if a.Copies != b.Copies {
			return a.Copies > b.Copies
		}
		return evolution.PolicyKey(a.Genome, a.Policy) < evolution.PolicyKey(b.Genome, b.Policy)
	})
	return r, nil
}
