package rules

import (
	"open-end/internal/vm"
	"testing"
)

type chemicalSink struct {
	transfers int
	reactions []ReactionEvent
}

func (*chemicalSink) Interaction(Interaction) {}
func (*chemicalSink) Death(Death)             {}
func (s *chemicalSink) ChemicalMoved(from, to, species, units, available int) {
	if units > 0 && units <= available && from != to {
		s.transfers++
	}
}
func (s *chemicalSink) Reacted(e ReactionEvent) { s.reactions = append(s.reactions, e) }
func TestChemicalCallbacksRecordResolvedReactionAndConservativeDiffusion(t *testing.T) {
	w := fixture(t)
	w.Config.ChemicalDiffusion = 1
	p := w.Particles[1]
	s := &chemicalSink{}
	ResolveObserved(w, Event{Actor: 1, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.CONVERT, A: 0, B: 4}}}, s)
	if len(s.reactions) != 1 || s.reactions[0].Position != p.Position || s.reactions[0].Before[0]-s.reactions[0].After[0] != 4 || s.reactions[0].Energy != 16 {
		t.Fatal("reaction event wrong")
	}
	before, _ := w.Totals()
	transportObserved(w, s)
	after, _ := w.Totals()
	if s.transfers == 0 || before != after {
		t.Fatal("transport not observed conservatively")
	}
	check(t, w)
}
