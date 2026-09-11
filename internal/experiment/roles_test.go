package experiment

import (
	"bytes"
	"open-end/internal/evolution"
	"open-end/internal/kernel"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"reflect"
	"testing"
)

func roleFixture(t *testing.T) *world.World {
	t.Helper()
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.MutationPPM = 0
	c.Inflow = 128
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	p := w.Particles[1]
	for _, i := range []vm.Instruction{{Op: vm.ALLOCATE, A: 1}, {Op: vm.COPY}, {Op: vm.TRANSFER, A: 32}, {Op: vm.BIND}} {
		rules.Resolve(w, rules.Event{Actor: p.ID, Intent: vm.Intent{Instruction: i}})
	}
	programs := [][]vm.Instruction{{{Op: vm.ABSORB, A: 32}, {Op: vm.TRANSFER, A: 16}}, {{Op: vm.NOP}}}
	for j, code := range programs {
		p := w.Particles[uint64(j+1)]
		p.Code = code
		p.IP = 0
		p.Origin = evolution.Hash(code, p.InitialMemory)
		w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
		w.RegisterGenome(p, "")
	}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
	return w
}
func TestRoleAssayDetectsKnownDependentPartner(t *testing.T) {
	w := roleFixture(t)
	sourceHash := kernel.Hash(w)
	p := RoleProtocol{Version: 1, Mode: "intact", Members: []uint64{1, 2}, Ticks: 100, Every: 10}
	a, control, err := ContinueRoles(w, p)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := kernel.Save(&buf, w); err != nil {
		t.Fatal(err)
	}
	baseline, _ := kernel.Load(&buf)
	for range p.Ticks {
		kernel.Step(baseline)
	}
	if kernel.Hash(a) != kernel.Hash(baseline) || kernel.Hash(w) != sourceHash {
		t.Fatal("observer changed physics or source")
	}
	p.Mode = "no-sharing"
	_, sharing, err := ContinueRoles(w, p)
	if err != nil {
		t.Fatal(err)
	}
	if control.Roles[1].Alive != 1 || sharing.Roles[1].Alive != 0 || sharing.Roles[0].Alive != 1 || len(control.Flows) != 1 || len(sharing.Flows) != 0 {
		t.Fatal("failed known dependency", control, sharing)
	}
	p.Mode = "no-peer-sharing"
	_, peers, err := ContinueRoles(w, p)
	if err != nil || peers.FinalHash != sharing.FinalHash {
		t.Fatal("peer transfer intervention missed dependent partner", err)
	}
	p.Mode = "no-acquisition"
	p.Donor = 1
	_, starved, err := ContinueRoles(w, p)
	if err != nil {
		t.Fatal(err)
	}
	if starved.Roles[0].Acquired != 0 || starved.Roles[1].FounderAliveTicks >= control.Roles[1].FounderAliveTicks {
		t.Fatal("donor effect on partner missed")
	}
	_, again, err := ContinueRoles(w, p)
	if err != nil || !reflect.DeepEqual(again, starved) {
		t.Fatal("protocol replay differs", err)
	}
	p.Mode = "no-signals"
	p.Donor = 0
	_, signals, err := ContinueRoles(w, p)
	if err != nil || signals.FinalHash != control.FinalHash {
		t.Fatal("unexercised treatment changed physics")
	}
	p.Mode = "no-bonds"
	_, bonds, err := ContinueRoles(w, p)
	if err != nil || bonds.RemovedBonds != 1 || bonds.ConnectedPairTicks != 0 {
		t.Fatal("bond treatment failed", err)
	}
}

func TestRoleTagsRequireActualCopyAndSurviveFounderDeath(t *testing.T) {
	w := roleFixture(t)
	s := &roleSink{w: w, tags: map[uint64]int{1: 1, 2: 2}, r: RoleResult{Protocol: RoleProtocol{Mode: "no-acquisition", Donor: 1}, Roles: []CellRole{{Founder: 1, Executed: map[vm.Opcode]uint64{}, Suppressed: map[vm.Opcode]uint64{}}, {Founder: 2}}}, flows: map[string]*RoleFlow{}}
	p := w.Particles[1]
	act := func(op vm.Opcode, a int) {
		rules.ResolveWithIntervention(w, rules.Event{Actor: 1, Intent: vm.Intent{Instruction: vm.Instruction{Op: op, A: a}}}, s, s.before)
	}
	act(vm.ALLOCATE, 0)
	id := p.Target
	if id == 0 || s.tags[id] != 0 {
		t.Fatal("allocation incorrectly tagged")
	}
	act(vm.COPY, 0)
	if s.tags[id] != 1 || s.r.Roles[0].Copies != 1 {
		t.Fatal("copy did not transmit tag")
	}
	s.Death(rules.Death{ID: 1})
	if s.tags[id] != 1 {
		t.Fatal("founder death erased descendant ancestry")
	}
	e := rules.Event{Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ABSORB, A: 32}}}
	if !s.before(w.Particles[id], &e) {
		t.Fatal("intervention did not follow COPY lineage")
	}
	// A peer-only ablation preserves the original cell's own copied offspring.
	s.r.Protocol.Mode = "no-peer-sharing"
	s.tags[1] = 1
	p.Target = id
	e.Intent.Op = vm.TRANSFER
	if s.before(p, &e) {
		t.Fatal("peer-only ablation blocks offspring provisioning")
	}
	p.Target = 2
	if !s.before(p, &e) {
		t.Fatal("peer-only ablation allows cross-lineage transfer")
	}
	p.Target = 999
	if s.before(p, &e) {
		t.Fatal("peer-only ablation targets outsiders")
	}
}

func TestRoleSignalInterventionBlindsOnlyCommunicationInputs(t *testing.T) {
	w := roleFixture(t)
	s := &roleSink{w: w, tags: map[uint64]int{1: 1}, r: RoleResult{Protocol: RoleProtocol{Mode: "no-signals"}, Roles: []CellRole{{Executed: map[vm.Opcode]uint64{}, Suppressed: map[vm.Opcode]uint64{}}}}}
	for _, a := range []int{0, 3, 6, 7, 8, 11, 12, 15} {
		e := rules.Event{Sensed: 42, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.SENSE, A: a}}}
		if s.before(w.Particles[1], &e) {
			t.Fatal("sense must still write its result")
		}
		want := 42
		if a == 7 || a >= 12 && a <= 15 {
			want = 0
		}
		if e.Sensed != want {
			t.Fatal("wrong sensory channel", a)
		}
	}
	e := rules.Event{Sensed: 42, WordLength: 2, ForeignWord: true, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.LISTEN}}}
	if s.before(w.Particles[1], &e) || e.Sensed != 0 || e.WordLength != 0 || e.ForeignWord {
		t.Fatal("symbol reading not blinded")
	}
	e.Intent.Op = vm.TOKEN
	if s.before(w.Particles[2], &e) || !s.before(w.Particles[1], &e) {
		t.Fatal("emission scope incorrect")
	}
}

func TestRoleProtocolRejectsInvalidMembers(t *testing.T) {
	w := roleFixture(t)
	for _, ids := range [][]uint64{{1}, {2, 1}, {1, 1}, {1, 99}} {
		if _, _, e := ContinueRoles(w, RoleProtocol{Version: 1, Mode: "intact", Members: ids, Ticks: 1, Every: 1}); e == nil {
			t.Fatal("accepted invalid group", ids)
		}
	}
}
