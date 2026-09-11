package experiment

import (
	"math"
	"open-end/internal/evolution"
	"open-end/internal/kernel"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func chemicalFixture(t *testing.T) *world.World {
	t.Helper()
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.MutationPPM = 0
	c.Inflow = 128
	c.Ecology = true
	c.ChemicalDiffusion = 4
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range []vm.Instruction{{Op: vm.ALLOCATE, A: 1}, {Op: vm.COPY}, {Op: vm.TRANSFER, A: 32}, {Op: vm.BIND}} {
		rules.Resolve(w, rules.Event{Actor: 1, Intent: vm.Intent{Instruction: i}})
	}
	for j := 0; j < 2; j++ {
		p := w.Particles[uint64(j+1)]
		p.Code = []vm.Instruction{{Op: vm.CONVERT, A: j, B: 8}}
		p.IP = 0
		p.Origin = evolution.Hash(p.Code, p.InitialMemory)
		w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
		w.RegisterGenome(p, "")
	}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestChemicalTraceFindsKnownCrossFeedingWithoutChangingPhysics(t *testing.T) {
	w := chemicalFixture(t)
	hash := kernel.Hash(w)
	p := RoleProtocol{Version: 2, Mode: "intact", Members: []uint64{1, 2}, Ticks: 100, Every: 20}
	_, r, err := ContinueRoles(w, p)
	if err != nil {
		t.Fatal(err)
	}
	p.Version = 1
	_, baseline, err := ContinueRoles(w, p)
	if err != nil {
		t.Fatal(err)
	}
	if r.FinalHash != baseline.FinalHash || kernel.Hash(w) != hash {
		t.Fatal("tracer altered physical replay")
	}
	c := r.Chemistry
	if c.Consumed[2][2] <= 0 || c.Roles[0].Units[1] != 0 || c.Roles[1].Units[0] != 0 {
		t.Fatal("known complementary metabolism not traced", c)
	}
	if math.Abs(c.Consumed[2][2]-float64(c.Roles[1].Units[1])) > 1e-8 {
		t.Fatal("consumer provenance attribution wrong")
	}
	p.Version = 2
	p.Mode = "no-reaction0"
	p.Donor = 1
	_, blocked, err := ContinueRoles(w, p)
	if err != nil {
		t.Fatal(err)
	}
	if blocked.Chemistry.Roles[0].Units[0] != 0 || blocked.Chemistry.Consumed[2][2] != 0 || blocked.Roles[1].FounderAliveTicks >= r.Roles[1].FounderAliveTicks {
		t.Fatal("blocked producer did not affect known dependent consumer")
	}
}

func TestTracerMixingAndUnknownPoolConserveFractionalMass(t *testing.T) {
	w := chemicalFixture(t)
	// A synthetic tracer-only pool checks the documented proportional convention.
	tc := newChemicalTrace(w, []uint64{1, 2})
	tc.labels[0][0] = 2
	tc.result.Initial[0] = 2
	tc.labels[0][2] = 6
	tc.result.Produced[2] = 6
	tc.move(0, 1, 3, 8)
	if tc.labels[1][0] != .75 || tc.labels[1][2] != 2.25 {
		t.Fatal("wrong mixture transported")
	}
	tc.react(rules.ReactionEvent{Actor: 2, Position: 1, Reaction: 1, Before: [3]int{0, 3, 0}, After: [3]int{0, 1, 2}}, 2)
	if tc.result.Consumed[0][2] != .5 || tc.result.Consumed[2][2] != 1.5 {
		t.Fatal("wrong mixture consumed")
	}
	// Endpoint chemistry is set solely for auditing this standalone tracer fixture.
	w.Cells[0].Chemical[1] = 5
	w.Cells[1].Chemical[1] = 1
	if err := tc.finish(w); err != nil {
		t.Fatal(err)
	}
}

func TestAnchoringSeparatesMotionFromInternalBonds(t *testing.T) {
	w := chemicalFixture(t)
	for _, p := range w.Particles {
		p.Code = append(p.Code, vm.Instruction{Op: vm.MOVE, A: -1})
		p.Origin = evolution.Hash(p.Code, p.InitialMemory)
		w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
		w.RegisterGenome(p, "")
	}
	p := RoleProtocol{Version: 2, Mode: "anchored", Members: []uint64{1, 2}, Ticks: 100, Every: 20}
	a, ar, err := ContinueRoles(w, p)
	if err != nil {
		t.Fatal(err)
	}
	p.Mode = "no-bonds-anchored"
	b, br, err := ContinueRoles(w, p)
	if err != nil {
		t.Fatal(err)
	}
	if br.RemovedBonds != 1 || ar.Roles[0].Suppressed[vm.MOVE] == 0 || br.Roles[0].Suppressed[vm.MOVE] == 0 {
		t.Fatal("anchoring treatment did not execute")
	}
	b.Relations = a.Relations
	if kernel.Hash(a) != kernel.Hash(b) {
		t.Fatal("fixed positions still depend on bonds in fixture")
	}
}
