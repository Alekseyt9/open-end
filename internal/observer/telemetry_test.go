package observer_test

import (
	"bytes"
	"encoding/json"
	"math"
	"open-end/internal/dsl"
	"open-end/internal/evolution"
	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/vm"
	"open-end/internal/world"
	"os"
	"testing"
)

func ecology(t *testing.T) *world.World {
	t.Helper()
	c := world.DefaultConfig()
	c.Width = 12
	c.Height = 12
	c.MaxEntities = 144
	c.Ecology = true
	c.MatterDiffusion = 4
	c.ChemicalDiffusion = 4
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	return w
}
func check(t *testing.T, w *world.World) {
	t.Helper()
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
}
func capture(t *testing.T, w *world.World, ticks, every int) []observer.Metrics {
	t.Helper()
	tr := observer.NewTracker(w)
	frames := []observer.Metrics{tr.Frame(w)}
	for i := 0; i < ticks; i++ {
		kernel.StepObserved(w, tr)
		if (i+1)%every == 0 {
			frames = append(frames, tr.Frame(w))
		}
	}
	check(t, w)
	return frames
}

func TestTelemetryLeavesPhysicsAndReplayUnchanged(t *testing.T) {
	a, b := ecology(t), ecology(t)
	tr := observer.NewTracker(b)
	tr.Frame(b)
	for i := 0; i < 600; i++ {
		kernel.Step(a)
		kernel.StepObserved(b, tr)
		if i%73 == 0 {
			tr.Frame(b)
		}
	}
	if kernel.Hash(a) != kernel.Hash(b) {
		t.Fatal("observer changed physics")
	}
	var saved bytes.Buffer
	if err := kernel.Save(&saved, b); err != nil {
		t.Fatal(err)
	}
	resumed, err := kernel.Load(&saved)
	if err != nil {
		t.Fatal(err)
	}
	frames := capture(t, resumed, 300, 100)
	for i := 0; i < 300; i++ {
		kernel.Step(a)
	}
	if kernel.Hash(a) != kernel.Hash(resumed) {
		t.Fatal("resumed observation changed replay")
	}
	r, err := observer.Summarize(frames, 300)
	if err != nil {
		t.Fatal(err)
	}
	if r.FromTick != 600 || r.ToTick != 900 || frames[0].Telemetry.SessionStart != 600 {
		t.Fatal("resume claimed earlier observation coverage")
	}
}

func TestLifetimesIncludeDeathsBetweenFrames(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 4
	c.Height = 4
	c.MaxEntities = 16
	c.CellCapacity = 1
	c.Inflow = 0
	c.Maintenance = 128
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	frames := capture(t, w, 5, 5)
	r, err := observer.Summarize(frames, 5)
	if err != nil {
		t.Fatal(err)
	}
	if r.Lifetimes.Count != 1 || r.Lifetimes.Mean != 1 || r.Lifetimes.Min != 1 || r.Lifetimes.Max != 1 || r.Population.End != 0 {
		t.Fatalf("wrong completed lifetime: %+v", r.Lifetimes)
	}
	if frames[1].Telemetry.AliveAges.Count != 0 {
		t.Fatal("dead particle counted as survivor")
	}
}

func register(w *world.World, p *world.Particle) {
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	if w.Origins[p.Origin] == nil {
		w.Origins[p.Origin] = &world.Origin{Hash: p.Origin, Copies: 1}
	}
	w.RegisterGenome(p, "")
}
func add(w *world.World, id uint64, pos, energy int, code []vm.Instruction) *world.Particle {
	p := &world.Particle{ID: id, Position: pos, Energy: energy, Code: code}
	if len(code) > 0 {
		register(w, p)
	}
	w.Particles[id] = p
	w.Cells[pos].Occupant = id
	w.Cells[pos].Matter--
	w.NextID = max(w.NextID, id+1)
	e, m := w.Totals()
	w.Accounting.InitialEnergy = e
	w.Accounting.InitialMatter = m
	return p
}

func TestStructureComponentsAndDiversity(t *testing.T) {
	w := ecology(t)
	p := w.Particles[1]
	east := w.Neighbor(p.Position, 1)
	south := w.Neighbor(east, 2)
	add(w, 2, east, 128, vm.Seed())
	add(w, 3, south, 128, vm.EcologySeed())
	add(w, 4, 0, 128, vm.Seed())
	w.Relations[world.RelationKey(1, 2)] = world.Relation{A: 1, B: 2}
	w.Relations[world.RelationKey(2, 3)] = world.Relation{A: 2, B: 3}
	check(t, w)
	before := kernel.Hash(w)
	tr := observer.NewTracker(w)
	m := tr.Frame(w)
	x := m.Telemetry
	if kernel.Hash(w) != before {
		t.Fatal("frame changed state")
	}
	if x.Structures.Components != 2 || x.Structures.LinkedComponents != 1 || x.Structures.LinkedParticles != 3 || x.Structures.Largest != 3 || x.Structures.MixedGenomeComponents != 1 {
		t.Fatalf("bad structures: %+v", x.Structures)
	}
	if len(x.Structures.Bindings) != 1 || x.Structures.Bindings[0].Count != 2 {
		t.Fatal("incorrect aggregated binding graph")
	}
	if math.Abs(x.Diversity.Shannon-math.Log(2)) > 1e-12 || math.Abs(x.Diversity.EffectiveGenomes-2) > 1e-12 || x.Diversity.DominantShare != 0.5 {
		t.Fatal("incorrect diversity")
	}
	a, _ := json.Marshal(m)
	b, _ := json.Marshal(tr.Frame(w))
	if !bytes.Equal(a, b) {
		t.Fatal("snapshot telemetry is not deterministic")
	}
}

func TestInteractionDirectionsAndOnlySuccessfulEvents(t *testing.T) {
	for _, op := range []vm.Opcode{vm.TRANSFER, vm.TAKE, vm.BIND, vm.UNBIND} {
		t.Run(string(rune('A'+op)), func(t *testing.T) {
			w := ecology(t)
			p := w.Particles[1]
			p.Code = []vm.Instruction{{Op: op, A: 5}}
			if op == vm.UNBIND {
				p.Code[0].A = -1
			}
			register(w, p)
			q := add(w, 2, w.Neighbor(p.Position, 1), 20, nil)
			p.Target = q.ID
			if op == vm.UNBIND {
				w.Relations[world.RelationKey(1, 2)] = world.Relation{A: 1, B: 2}
			}
			frames := capture(t, w, 1, 1)
			r, err := observer.Summarize(frames, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(r.Interactions) != 1 {
				t.Fatalf("missing action: %+v", r.Interactions)
			}
			edge := r.Interactions[0]
			if op == vm.TAKE {
				if edge.Source != "" || edge.Target != p.Genome || edge.Energy != 5 {
					t.Fatal("TAKE direction incorrect")
				}
			} else if edge.Source != p.Genome || edge.Target != "" {
				t.Fatal("actor direction incorrect")
			}
		})
	}
	w := ecology(t)
	p := w.Particles[1]
	p.Code = []vm.Instruction{{Op: vm.TRANSFER, A: 5}}
	p.Target = 999
	register(w, p)
	frames := capture(t, w, 1, 1)
	if len(frames[1].Telemetry.Interactions) != 0 {
		t.Fatal("failed transfer produced an edge")
	}
}

func TestSummaryReconcilesTransientGenomesAndRejectsGaps(t *testing.T) {
	w := ecology(t)
	w.Config.MutationPPM = 100000
	frames := capture(t, w, 1000, 100)
	r, err := observer.Summarize(frames, 850)
	if err != nil {
		t.Fatal(err)
	}
	if r.FromTick != 100 || r.ToTick != 1000 || len(r.Discoveries) == 0 || r.Copies == 0 || r.Lifetimes.Count == 0 {
		t.Fatal("window or activity missing")
	}
	for _, g := range r.Activity {
		if int64(g.End) != int64(g.Start)+int64(g.Births)-int64(g.Deaths) {
			t.Fatal("genotype balance differs")
		}
	}
	gap := append([]observer.Metrics{frames[0]}, frames[2:]...)
	if _, err := observer.Summarize(gap, 1000); err == nil {
		t.Fatal("accepted missing frame")
	}
	bad := append([]observer.Metrics{}, frames...)
	x := *bad[2].Telemetry
	x.Lifetimes.Count++
	bad[2].Telemetry = &x
	if _, err := observer.Summarize(bad, 1000); err == nil {
		t.Fatal("accepted corrupt death count")
	}
	legacy := append([]observer.Metrics{}, frames...)
	legacy[0].Telemetry = nil
	if _, err := observer.Summarize(legacy, 1000); err == nil {
		t.Fatal("treated legacy metrics as exact telemetry")
	}
	mixed := append([]observer.Metrics{}, frames...)
	other := *mixed[2].Telemetry
	other.SessionHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	mixed[2].Telemetry = &other
	if _, err := observer.Summarize(mixed, 1000); err == nil {
		t.Fatal("accepted mixed source worlds")
	}
	missingEdge := append([]observer.Metrics{}, frames...)
	other = *missingEdge[2].Telemetry
	other.Interactions = []observer.Edge{}
	missingEdge[2].Telemetry = &other
	if _, err := observer.Summarize(missingEdge, 1000); err == nil {
		t.Fatal("accepted missing interaction events")
	}
	tr := observer.NewTracker(w)
	tr.Frame(w)
	kernel.Step(w)
	if tr.Frame(w).Telemetry.Complete {
		t.Fatal("unobserved tick reported complete")
	}
}

func TestTelemetryKeepsRuleTransitionsInsideWindow(t *testing.T) {
	w := ecology(t)
	load := func(name string) *dsl.Module {
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
	base, changed := load("baseline"), load("direct-x")
	if err := kernel.ReloadRules(w, base); err != nil {
		t.Fatal(err)
	}
	if err := kernel.ScheduleRules(w, dsl.Change{Tick: 100, Module: changed}); err != nil {
		t.Fatal(err)
	}
	if err := kernel.ScheduleRules(w, dsl.Change{Tick: 200, Rollback: true}); err != nil {
		t.Fatal(err)
	}
	r, err := observer.Summarize(capture(t, w, 300, 50), 300)
	if err != nil {
		t.Fatal(err)
	}
	if r.RulesStart.Hash != base.Hash || r.RulesEnd.Hash != base.Hash || len(r.RuleEvents) != 2 {
		t.Fatal("lost intermediate rule changes")
	}
	if r.Flows.DSL[changed.Hash+"/x-to-z"] == 0 {
		t.Fatal("missing versioned reaction usage")
	}
}
