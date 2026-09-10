package kernel

import (
	"bytes"
	"encoding/json"
	"open-end/internal/dsl"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"os"
	"testing"
)

func ruleModule(t *testing.T, name string) *dsl.Module {
	t.Helper()
	f, err := os.Open("../../examples/rules/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := dsl.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func ruleWorld(t *testing.T) *world.World {
	t.Helper()
	c := world.DefaultConfig()
	c.Width = 12
	c.Height = 12
	c.MaxEntities = 144
	c.Ecology = true
	c.MatterDiffusion = 4
	c.ChemicalDiffusion = 4
	c.MutationPPM = 0
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestHotReloadRollbackAndSnapshot(t *testing.T) {
	w := ruleWorld(t)
	base := ruleModule(t, "baseline")
	changed := ruleModule(t, "direct-x")
	if err := ReloadRules(w, base); err != nil {
		t.Fatal(err)
	}
	if err := ScheduleRules(w, dsl.Change{Tick: 100, Module: changed}); err != nil {
		t.Fatal(err)
	}
	if err := ScheduleRules(w, dsl.Change{Tick: 200, Rollback: true}); err != nil {
		t.Fatal(err)
	}
	// Installation and scheduling own their input, including bytecode and maps.
	changed.Rules[0].Code[0].Amount = 64
	changed.Source.Rules[0].Consume["X"] = 64
	advance(t, w, 150)
	if w.RuleState.Active.Hash == base.Hash || w.RuleState.Instructions == 0 {
		t.Fatal("new rules did not execute")
	}
	key := w.RuleState.Active.Hash + "/x-to-z"
	if w.RuleState.Usage[key] == 0 {
		t.Fatal("replacement reaction unused")
	}
	var buf bytes.Buffer
	if err := Save(&buf, w); err != nil {
		t.Fatal(err)
	}
	saved := append([]byte(nil), buf.Bytes()...)
	resumed, err := Load(&buf)
	if err != nil {
		t.Fatal(err)
	}
	advance(t, w, 150)
	advance(t, resumed, 150)
	if Hash(w) != Hash(resumed) || w.RuleState.Active.Hash != base.Hash || len(w.RuleState.Events) != 3 {
		t.Fatal("scheduled rollback replay differs")
	}
	var s snapshot
	if err := json.Unmarshal(saved, &s); err != nil {
		t.Fatal(err)
	}
	if s.Format != 3 {
		t.Fatal("wrong snapshot format")
	}
	s.World.RuleState.Active.Rules[0].Code[0].Amount++
	corrupt, _ := json.Marshal(s)
	if _, err := Load(bytes.NewReader(corrupt)); err == nil {
		t.Fatal("loaded corrupt rule bytecode")
	}
}

func TestRuleInstallFailureLeavesStateUnchanged(t *testing.T) {
	w := ruleWorld(t)
	before := Hash(w)
	if err := RollbackRules(w); err == nil || Hash(w) != before {
		t.Fatal("empty rollback mutated world")
	}
	m := ruleModule(t, "baseline")
	m.Rules[0].EnergyCost = 0
	if err := ReloadRules(w, m); err == nil || Hash(w) != before {
		t.Fatal("bad reload mutated world")
	}
	m = ruleModule(t, "baseline")
	if err := ReloadRules(w, m); err != nil {
		t.Fatal(err)
	}
	if err := ScheduleRules(w, dsl.Change{Tick: 100, Rollback: true}); err != nil {
		t.Fatal(err)
	}
	before = Hash(w)
	if err := RollbackRules(w); err == nil || Hash(w) != before {
		t.Fatal("invalidated pending rollback")
	}
	if err := ScheduleRules(w, dsl.Change{Tick: 100, Module: m}); err == nil || Hash(w) != before {
		t.Fatal("duplicate tick accepted")
	}
}

func TestBaselineDSLMatchesBuiltinPhysics(t *testing.T) {
	a, b := ruleWorld(t), ruleWorld(t)
	if err := ReloadRules(b, ruleModule(t, "baseline")); err != nil {
		t.Fatal(err)
	}
	advance(t, a, 300)
	advance(t, b, 300)
	b.RuleState = nil // Only the audit metadata may differ.
	if Hash(a) != Hash(b) {
		t.Fatal("baseline DSL changed physical trajectory")
	}
	if envelope(a).Format != 2 {
		t.Fatal("legacy worlds changed format")
	}
}

func TestImmediateRuleRollback(t *testing.T) {
	w := ruleWorld(t)
	advance(t, w, 20)
	tick := w.Tick
	if err := ReloadRules(w, ruleModule(t, "direct-x")); err != nil {
		t.Fatal(err)
	}
	advance(t, w, 20)
	before := w.Accounting
	if err := RollbackRules(w); err != nil {
		t.Fatal(err)
	}
	if w.Tick != tick+20 || w.RuleState.Active != nil || w.Accounting != before {
		t.Fatal("rollback reset physical state")
	}
	advance(t, w, 20)
}

func TestNewReactionIDAndCustomAttemptCost(t *testing.T) {
	w := ruleWorld(t)
	m := ruleModule(t, "direct-x")
	m.Source.Rules[1].EnergyCost = 3
	m, err := dsl.Compile(m.Source)
	if err != nil {
		t.Fatal(err)
	}
	if err := ReloadRules(w, m); err != nil {
		t.Fatal(err)
	}
	p := w.Particles[1]
	// First create Z using the replacement rule, then charge it via new ID 2.
	rules.Resolve(w, rules.Event{Actor: p.ID, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.CONVERT, A: 0, B: 1}}})
	energy, chemical := p.Energy, w.Cells[p.Position].Chemical
	rules.Resolve(w, rules.Event{Actor: p.ID, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.CONVERT, A: 2, B: 1}}})
	if p.Energy != energy-11 || w.Cells[p.Position].Chemical != [3]int{chemical[0] + 1, chemical[1], chemical[2] - 1} {
		t.Fatal("new ID or custom fee did not apply")
	}
	if w.RuleState.Usage[m.Hash+"/charge-x"] != 1 {
		t.Fatal("new rule usage missing")
	}
	energy = p.Energy
	rules.Resolve(w, rules.Event{Actor: p.ID, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.CONVERT, A: 2, B: 1}}})
	if p.Energy != energy-3 {
		t.Fatal("failed reaction did not pay attempt cost")
	}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
}
