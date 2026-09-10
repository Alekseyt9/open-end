package world

import (
	"fmt"
	"open-end/internal/evolution"
	"open-end/internal/vm"
)

func (w *World) RecordCopy(genome string, policy evolution.CopyPolicy, changed bool, donor string) {
	if w.Variation == nil {
		w.Variation = map[string]*evolution.CopyRecord{}
	}
	key := evolution.PolicyKey(genome, policy)
	rec := w.Variation[key]
	if rec == nil {
		rec = &evolution.CopyRecord{Genome: genome, Policy: policy, FirstTick: w.Tick}
		w.Variation[key] = rec
	}
	rec.Copies++
	rec.LastTick = w.Tick
	if changed {
		rec.Changed++
	}
	if donor != "" {
		rec.Recombined++
		if rec.Donors == nil {
			rec.Donors = map[string]uint64{}
		}
		rec.Donors[donor]++
	}
}
func (w *World) ValidateVariation() error {
	if w.Config.CopyModel == "" {
		if len(w.Variation) > 0 {
			return fmt.Errorf("variation ledger in legacy world")
		}
		return nil
	}
	counts := map[string]uint64{}
	var copies uint64
	for key, r := range w.Variation {
		if r == nil || !r.Policy.Valid() || key != evolution.PolicyKey(r.Genome, r.Policy) || r.Copies == 0 || r.Changed > r.Copies || r.Recombined > r.Copies || r.FirstTick > r.LastTick || r.LastTick > w.Tick || w.Genomes[r.Genome] == nil {
			return fmt.Errorf("invalid variation record")
		}
		available := false
		for _, i := range w.Genomes[r.Genome].Code {
			if i.Op == vm.COPY && r.Policy.Kind == "code" || i.Op == vm.COPYMEM && r.Policy.Kind == "memory" {
				if evolution.DecodePolicy(i, w.Config.MutationPPM, w.Config.CopyModel == "evolving") == r.Policy {
					available = true
					break
				}
			}
		}
		if !available {
			return fmt.Errorf("variation policy is not encoded by its genome")
		}
		var donors uint64
		for hash, n := range r.Donors {
			if w.Genomes[hash] == nil || n == 0 || n > r.Recombined-donors {
				return fmt.Errorf("invalid donor ledger")
			}
			donors += n
		}
		if donors != r.Recombined || r.Policy.Kind == "memory" && donors != 0 || donors > 0 && !r.Policy.Recombine {
			return fmt.Errorf("invalid recombination counts")
		}
		if r.Policy.Kind == "code" {
			if r.Copies > w.Accounting.Copies-copies {
				return fmt.Errorf("copy policy count overflow")
			}
			copies += r.Copies
			counts[r.Genome] += r.Copies
		}
	}
	if copies != w.Accounting.Copies {
		return fmt.Errorf("copy policy ledger does not reconcile")
	}
	for hash, g := range w.Genomes {
		if g != nil && counts[hash] != g.Copies {
			return fmt.Errorf("genome copy policy counts disagree")
		}
	}
	return nil
}
