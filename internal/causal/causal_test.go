package causal

import (
	"fmt"
	"math"
	"open-end/internal/discovery"
	"open-end/internal/evolution"
	"open-end/internal/kernel"
	"open-end/internal/vm"
	"open-end/internal/world"
	"reflect"
	"slices"
	"testing"
)

func fixture(t *testing.T) (*world.World, discovery.Report) {
	t.Helper()
	c := world.DefaultConfig()
	c.Width, c.Height, c.MaxEntities = 8, 8, 64
	c.MutationPPM = 0
	w, e := world.New(c)
	if e != nil {
		t.Fatal(e)
	}
	p := w.Particles[1]
	w.Cells[p.Position].Occupant = 0
	w.Cells[p.Position].Matter++
	p.Position = 9
	w.Cells[9].Matter--
	w.Cells[9].Occupant = 1
	p.Code = []vm.Instruction{{Op: vm.NOP}}
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
	w.RegisterGenome(p, "")
	for _, pos := range []int{10, 17, 18} {
		q := &world.Particle{ID: w.NextID, Position: pos, Energy: 16, Code: slices.Clone(p.Code), Origin: p.Origin}
		w.NextID++
		w.Cells[pos].Energy -= 16
		w.Cells[pos].Matter--
		w.Cells[pos].Occupant = q.ID
		w.Particles[q.ID] = q
		w.RegisterGenome(q, "")
	}
	w.Relations[world.RelationKey(1, 2)] = world.Relation{A: 1, B: 2}
	w.Relations[world.RelationKey(3, 4)] = world.Relation{A: 3, B: 4}
	w.Tick = 100
	if e := w.Validate(); e != nil {
		t.Fatal(e)
	}
	nodes := []discovery.Node{}
	for _, ids := range [][]uint64{{1, 2}, {3, 4}} {
		nodes = append(nodes, discovery.Node{ID: identity(ids), Members: ids, Sources: []string{"bonds"}, PersistentBond: true, BondAge: 100})
	}
	r := discovery.Report{Version: 1, Treatment: "intact", Config: discovery.Config{MinAge: 10}, FinalHash: kernel.Hash(w), Frames: []discovery.Frame{{Tick: w.Tick, MicroIDs: []uint64{1, 2, 3, 4}, Nodes: nodes}}}
	return w, r
}
func TestSelectionAndPermutationsDoNotChangeWorldOrMarginals(t *testing.T) {
	w, r := fixture(t)
	before := kernel.Hash(w)
	c := DefaultConfig()
	g, e := Select(w, r, c)
	if e != nil {
		t.Fatal(e)
	}
	if len(g) != 8 || kernel.Hash(w) != before {
		t.Fatal("selection changed physics")
	}
	for part := 0; part < 4; part++ {
		ids := []uint64{}
		for _, v := range g[part*2 : part*2+2] {
			if len(v.Members) != 2 {
				t.Fatal("changed sizes")
			}
			ids = append(ids, v.Members...)
		}
		slices.Sort(ids)
		if !slices.Equal(ids, []uint64{1, 2, 3, 4}) {
			t.Fatal("changed pool")
		}
	}
	copy, e := Select(w, r, c)
	if e != nil || !reflect.DeepEqual(g, copy) {
		t.Fatal("nondeterministic partition")
	}
	r.Frames[0].Nodes[0].Members = []uint64{1, 3}
	r.Frames[0].Nodes[0].ID = identity([]uint64{1, 3})
	if _, e := Select(w, r, c); e == nil {
		t.Fatal("disconnected selected boundary accepted")
	}
}
func TestMatchedPulseInterventionsConserveAndPermitUnavailableControl(t *testing.T) {
	w, r := fixture(t)
	c := DefaultConfig()
	c.Horizon, c.Blocks = 3, 2
	groups, e := Select(w, r, c)
	if e != nil {
		t.Fatal(e)
	}
	g := groups[0]
	hash := kernel.Hash(w)
	baseline, e := Forecast(w, groups, c)
	if e != nil {
		t.Fatal(e)
	}
	plain, e := clone(w)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < c.Horizon; i++ {
		kernel.Step(plain)
	}
	if kernel.Hash(plain) != baseline.FirstHash {
		t.Fatal("control counter changed physical state")
	}
	for _, mode := range []string{"local-cut", "outside-cut"} {
		v, e := Intervene(w, g, mode, 3)
		if e != nil {
			t.Fatal(e)
		}
		if !v.Available || len(v.Removed) != 1 || v.Outcome.Members != 2 || v.SourceHash != hash {
			t.Fatal("unmatched intervention")
		}
		target := slices.Contains(g.Members, v.Removed[0].A)
		if target != (mode == "local-cut") {
			t.Fatal("wrong cut location")
		}
	}
	if kernel.Hash(w) != hash {
		t.Fatal("source modified")
	}
	for key, b := range w.Relations {
		if !slices.Contains(g.Members, b.A) {
			delete(w.Relations, key)
		}
	}
	v, e := Intervene(w, g, "outside-cut", 3)
	if e != nil || v.Available || v.Reason == "" || v.FinalHash != "" {
		t.Fatal("invented unavailable control")
	}
}
func TestForecastFreezesSurvivorsAndNeverAddsNewMembers(t *testing.T) {
	w, _ := fixture(t)
	p := w.Particles[2]
	w.Accounting.Dissipated += int64(p.Energy - 2)
	p.Energy = 2
	g := Group{identity([]uint64{1, 2}), "bonds", []uint64{1, 2}}
	c := DefaultConfig()
	c.Horizon, c.Blocks = 3, 2
	r, e := Forecast(w, []Group{g}, c)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Samples) != 1 || r.Samples[0].Y != 0.5 || r.Samples[0].Members[1].Alive != 0 {
		t.Fatal("target or eligibility uses future membership", r.Samples)
	}
}
func TestRidgeAnalyticSolutionAndNonfiniteRejection(t *testing.T) {
	r, e := fit([]row{{[]float64{0}, 0, .5}, {[]float64{1}, 1, .5}}, .01)
	if e != nil {
		t.Fatal(e)
	}
	if math.Abs(r.weights[1]-.25/.26) > 1e-10 || math.Abs(r.weights[0]-(1-r.weights[1])/2) > 1e-10 {
		t.Fatal("incorrect ridge solution", r.weights)
	}
	if _, e := fit([]row{{[]float64{math.NaN()}, 0, 1}}, .01); e == nil {
		t.Fatal("nonfinite accepted")
	}
}
func examples() []Sample {
	ss := []Sample{}
	for seed := uint64(1); seed <= 4; seed++ {
		for group := 0; group < 4; group++ {
			y := float64(group % 2)
			x := make([]float64, len(MacroFeatures))
			x[len(MicroFeatures)] = y
			members := []Member{{1, make([]float64, len(MicroFeatures)), y}, {2, make([]float64, len(MicroFeatures)), y}}
			ss = append(ss, Sample{Seed: seed, Partition: "bonds", Group: fmt.Sprint(group), From: 100, To: 200, X: x, Members: members, Y: y})
		}
	}
	return ss
}
func TestHeldOutSeedsHaveNoTargetLeakageAndContextCanExplainGain(t *testing.T) {
	samples := examples()
	a, e := Evaluate(samples, .01)
	if e != nil {
		t.Fatal(e)
	}
	if a.ByPartition["bonds"]["micro"].MSE < .24 || a.ByPartition["bonds"]["macro"].MSE > .01 || a.ByPartition["bonds"]["micro-context"].MSE > .01 {
		t.Fatal("models fail known context example")
	}
	for _, f := range a.Folds {
		if slices.Contains(f.TrainSeeds, f.Seed) {
			t.Fatal("held-out world used in training")
		}
	}
	for i := range samples {
		if samples[i].Seed == 1 {
			samples[i].Y = 1 - samples[i].Y
			for j := range samples[i].Members {
				samples[i].Members[j].Alive = 1 - samples[i].Members[j].Alive
			}
		}
	}
	b, e := Evaluate(samples, .01)
	if e != nil {
		t.Fatal(e)
	}
	for i, p := range a.Predictions {
		if p.Seed == 1 && !reflect.DeepEqual(p.Values, b.Predictions[i].Values) {
			t.Fatal("held-out labels affected their own predictions")
		}
	}
}
