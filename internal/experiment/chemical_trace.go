package experiment

import (
	"fmt"
	"math"
	"open-end/internal/rules"
	"open-end/internal/world"
)

type ChemicalRole struct {
	Founder       uint64   `json:"founder"`
	Units         [2]int64 `json:"lineage_reaction_units"`
	OriginalUnits [2]int64 `json:"original_cell_reaction_units"`
}
type ChemicalTraceResult struct {
	Convention string         `json:"convention"`
	Roles      []ChemicalRole `json:"roles"`
	// Producer rows: initial unknown, outside production, then original member
	// lineages. Consumer columns: outside, then original member lineages.
	Initial         []float64   `json:"initial_y_by_producer"`
	Produced        []float64   `json:"produced_y_by_producer"`
	Consumed        [][]float64 `json:"consumed_y_by_producer_consumer"`
	Remaining       []float64   `json:"remaining_y_by_producer"`
	MaxCellError    float64     `json:"maximum_endpoint_cell_mass_error"`
	MaxBalanceError float64     `json:"maximum_producer_balance_error"`
}
type chemicalTrace struct {
	labels [][]float64
	result ChemicalTraceResult
}

func newChemicalTrace(w *world.World, members []uint64) *chemicalTrace {
	n := len(members) + 2
	t := &chemicalTrace{labels: make([][]float64, len(w.Cells)), result: ChemicalTraceResult{Convention: "Proportional well-mixed Y tracer; producer rows [initial-unknown,outside,member lineages]; consumer columns [outside,member lineages]. Labels track the most recent X->Y producer, not molecule identities.", Initial: make([]float64, n), Produced: make([]float64, n), Consumed: make([][]float64, n), Remaining: make([]float64, n)}}
	for i := range t.result.Consumed {
		t.result.Consumed[i] = make([]float64, len(members)+1)
	}
	for _, id := range members {
		t.result.Roles = append(t.result.Roles, ChemicalRole{Founder: id})
	}
	for pos, c := range w.Cells {
		t.labels[pos] = make([]float64, n)
		t.labels[pos][0] = float64(c.Chemical[1])
		t.result.Initial[0] += float64(c.Chemical[1])
	}
	return t
}
func (s *roleSink) ChemicalMoved(from, to, species, units, available int) {
	if s.trace == nil || species != 1 {
		return
	}
	s.trace.move(from, to, units, available)
}
func (t *chemicalTrace) move(from, to, units, available int) {
	if units <= 0 {
		return
	}
	fraction := float64(units) / float64(available)
	for i, amount := range t.labels[from] {
		n := amount * fraction
		t.labels[from][i] -= n
		t.labels[to][i] += n
	}
}
func (s *roleSink) Reacted(e rules.ReactionEvent) {
	if s.trace == nil {
		return
	}
	s.trace.react(e, s.tags[e.Actor])
}
func (t *chemicalTrace) react(e rules.ReactionEvent, lineage int) {
	if e.Reaction < 0 || e.Reaction > 1 {
		return
	}
	units := e.Before[e.Reaction] - e.After[e.Reaction]
	if units <= 0 {
		return
	}
	if lineage != 0 {
		r := &t.result.Roles[lineage-1]
		r.Units[e.Reaction] += int64(units)
		if e.Actor == r.Founder {
			r.OriginalUnits[e.Reaction] += int64(units)
		}
	}
	if e.Reaction == 0 {
		producer := lineage + 1 // zero lineage is outside production, not initial Y
		t.labels[e.Position][producer] += float64(units)
		t.result.Produced[producer] += float64(units)
	} else {
		fraction := float64(units) / float64(e.Before[1])
		for producer, amount := range t.labels[e.Position] {
			n := amount * fraction
			t.labels[e.Position][producer] -= n
			t.result.Consumed[producer][lineage] += n
		}
	}
}
func (t *chemicalTrace) finish(w *world.World) error {
	for pos, labels := range t.labels {
		sum := 0.0
		for i, amount := range labels {
			if math.IsNaN(amount) || math.IsInf(amount, 0) || amount < -1e-9 {
				return fmt.Errorf("invalid chemical tracer mass")
			}
			sum += amount
			t.result.Remaining[i] += amount
		}
		err := math.Abs(sum - float64(w.Cells[pos].Chemical[1]))
		t.result.MaxCellError = max(t.result.MaxCellError, err)
		if err > 1e-7 {
			return fmt.Errorf("chemical tracer drift in cell %d: %g", pos, err)
		}
	}
	for i, produced := range t.result.Produced {
		used := 0.0
		for _, amount := range t.result.Consumed[i] {
			used += amount
		}
		err := math.Abs(t.result.Initial[i] + produced - used - t.result.Remaining[i])
		t.result.MaxBalanceError = max(t.result.MaxBalanceError, err)
		if err > 1e-6+1e-10*produced {
			return fmt.Errorf("chemical tracer producer balance drift: %g", err)
		}
	}
	return nil
}
