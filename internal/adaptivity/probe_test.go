package adaptivity

import (
	"math"
	"open-end/internal/evolution"
	"open-end/internal/kernel"
	"open-end/internal/vm"
	"open-end/internal/world"
	"reflect"
	"strings"
	"testing"
)

func fixture(t *testing.T) *world.World {
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.MutationPPM = 0
	w, e := world.New(c)
	if e != nil {
		t.Fatal(e)
	}
	return w
}
func TestProbeSourceReplayAndUnexercisedPerception(t *testing.T) {
	w := fixture(t)
	hash := kernel.Hash(w)
	c := Config{90, .01}
	for _, ch := range Challenges() {
		a, e := ProbeWorld(w, c, ch, "intact")
		if e != nil {
			t.Fatal(e)
		}
		b, e := ProbeWorld(w, c, ch, "frozen-perception")
		if e != nil {
			t.Fatal(e)
		}
		if a.FinalHash != b.FinalHash || b.ChangedPerceptions != 0 {
			t.Fatal("seed without external sensing changed")
		}
		r, e := ProbeWorld(w, c, ch, "intact")
		if e != nil || !reflect.DeepEqual(r, a) {
			t.Fatal("probe not reproducible")
		}
		if kernel.Hash(w) != hash {
			t.Fatal("source mutated")
		}
	}
	normal, _ := Clone(w)
	filtered, _ := Clone(w)
	for i := 0; i < 90; i++ {
		kernel.Step(normal)
		kernel.StepWithPerception(filtered, nil)
	}
	if kernel.Hash(normal) != kernel.Hash(filtered) {
		t.Fatal("nil filter changes physics")
	}
}
func TestFrozenPerceptionActuallySuppressesUpdatedValues(t *testing.T) {
	w := fixture(t)
	p := w.Particles[1]
	p.Code = []vm.Instruction{{Op: vm.SENSE, A: 1, B: 0}, {Op: vm.ABSORB, A: 32}, {Op: vm.JUMP, A: 0}}
	w.RegisterGenome(p, "")
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
	r, e := ProbeWorld(w, Config{90, .01}, Challenges()[0], "frozen-perception")
	if e != nil {
		t.Fatal(e)
	}
	if r.ChangedPerceptions == 0 {
		t.Fatal("perception was not frozen")
	}
	p.Memory[3] = 123
	r, e = ProbeWorld(w, Config{90, .01}, Challenges()[0], "memory-reset")
	if e != nil || r.ResetParticles != 1 {
		t.Fatalf("memory reset not recorded: %+v, %v", r, e)
	}
}
func synthetic(c Config) []Probe {
	ps := []Probe{}
	for _, ch := range Challenges() {
		for i, m := range Modes {
			ps = append(ps, Probe{Challenge: ch.Name, Split: ch.Split, Mode: m, SourceHash: strings.Repeat("a", 64), Ticks: c.Ticks, InitialExecutable: 10, PopulationRetention: []float64{.9, .6, .7}[i], Copies: 1})
		}
	}
	return ps
}
func TestScoreSignedContrastsThresholdAndIntegrity(t *testing.T) {
	c := DefaultConfig()
	ps := synthetic(c)
	s, e := ScoreProbes(ps, "selection", c)
	if e != nil || math.Abs(s.AdaptiveProxy-22.5) > 1e-12 {
		t.Fatalf("wrong score %+v, %v", s, e)
	}
	ps[1].PopulationRetention = .99
	ps[2].PopulationRetention = .99
	ps[4].PopulationRetention = .99
	ps[5].PopulationRetention = .99
	s, e = ScoreProbes(ps, "selection", c)
	if e != nil || s.AdaptiveProxy != 0 || s.PerceptionBenefit >= 0 {
		t.Fatal("harm rewarded or signed contrast hidden")
	}
	ps = synthetic(c)
	c.MinBenefit = .3
	s, e = ScoreProbes(ps, "selection", c)
	if e != nil || s.AdaptiveProxy != 0 {
		t.Fatal("threshold ignored")
	}
	if _, e := ScoreProbes(append(ps, ps[0]), "selection", c); e == nil {
		t.Fatal("duplicate accepted")
	}
	ps[1].SourceHash = "other"
	if _, e := ScoreProbes(ps, "selection", c); e == nil {
		t.Fatal("unmatched arms accepted")
	}
}
func TestSelectionReservesDiversityExplorationAndIgnoresValidation(t *testing.T) {
	cs := make([]Candidate, 6)
	for i := range cs {
		cs[i] = Candidate{SourceHash: strings.Repeat(string(rune('a'+i)), 64), Selection: Score{AdaptiveProxy: float64(i)}, FunctionalDiversity: float64(i + 1), Descriptor: []float64{float64(i), 0, 0, 0, 0, 0, 0}}
	}
	if e := Select(cs, 4, 7); e != nil {
		t.Fatal(e)
	}
	reasons := map[string]int{}
	for _, c := range cs {
		if c.Selected {
			reasons[c.Reason]++
		}
	}
	if reasons["adaptive_proxy"] != 2 || reasons["behavioral_diversity"] != 1 || reasons["exploration"] != 1 {
		t.Fatal("selection budget", reasons)
	}
	before := append([]Candidate(nil), cs...)
	for i := range cs {
		cs[i].Validation = Score{AdaptiveProxy: 100}
		cs[i].Probes = []Probe{{PopulationRetention: 1}}
	}
	if e := Select(cs, 4, 7); e != nil {
		t.Fatal(e)
	}
	for i := range cs {
		if cs[i].Selected != before[i].Selected || cs[i].Reason != before[i].Reason {
			t.Fatal("validation leaked into selection")
		}
		cs[i].Selection.AdaptiveProxy = 0
	}
	if e := Select(cs, 4, 7); e != nil {
		t.Fatal(e)
	}
	for _, c := range cs {
		if c.Reason == "adaptive_proxy" {
			t.Fatal("zero evidence called intelligence")
		}
	}
}
func TestFunctionalProfilesAreDeterministicAndIgnoreUnusedLength(t *testing.T) {
	w := fixture(t)
	end, _ := Clone(w)
	for i := 0; i < 90; i++ {
		kernel.Step(end)
	}
	a, x := FunctionalDiversity(w, end)
	b, y := FunctionalDiversity(w, end)
	if a < 1 || a != b || !reflect.DeepEqual(x, y) {
		t.Fatal("unstable profile")
	}
	for _, g := range end.Genomes {
		g.Code = append(g.Code, vm.Instruction{Op: vm.NOP})
	}
	b, y = FunctionalDiversity(w, end)
	if a != b || !reflect.DeepEqual(x, y) {
		t.Fatal("unused code length rewarded")
	}
}
