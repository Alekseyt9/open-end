package kernel

import (
	"bytes"
	"encoding/json"
	"open-end/internal/evolution"
	"open-end/internal/observer"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func small(t testing.TB, seed uint64) *world.World {
	t.Helper()
	c := world.DefaultConfig()
	c.Width = 24
	c.Height = 24
	c.MaxEntities = 576
	c.Seed = seed
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func advance(t testing.TB, w *world.World, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		Step(w)
		if err := w.Validate(); err != nil {
			t.Fatalf("tick %d: %v", w.Tick, err)
		}
	}
}

func TestReplay(t *testing.T) {
	a := small(t, 42)
	b := small(t, 42)
	advance(t, a, 250)
	advance(t, b, 250)
	if Hash(a) != Hash(b) {
		t.Fatal("same seed differs")
	}
	var buf bytes.Buffer
	if err := Save(&buf, a); err != nil {
		t.Fatal(err)
	}
	c, err := Load(&buf)
	if err != nil {
		t.Fatal(err)
	}
	advance(t, a, 500)
	advance(t, c, 500)
	if Hash(a) != Hash(c) {
		t.Fatal("snapshot continuation differs")
	}
	other := small(t, 43)
	advance(t, other, 750)
	if Hash(a) == Hash(other) {
		t.Fatal("seed has no effect")
	}
}

func TestReplicationMutationAndAccounting(t *testing.T) {
	w := small(t, 7)
	w.Config.MutationPPM = 100000
	advance(t, w, 3000)
	m := observer.Observe(w)
	if m.Copies < 20 || m.Executable < 2 || m.Genomes < 2 || m.Lineages < 2 || m.Deaths == 0 {
		t.Fatalf("missing evolution: %+v", m)
	}
	for _, p := range w.Particles {
		if p.Origin != "" && w.Origins[p.Origin] == nil {
			t.Fatal("missing ancestry")
		}
	}
}

func TestZeroMutationPreservesGenome(t *testing.T) {
	w := small(t, 3)
	w.Config.MutationPPM = 0
	advance(t, w, 1200)
	m := observer.Observe(w)
	if m.Copies == 0 || m.Genomes != 1 {
		t.Fatalf("genetic control failed: %+v", m)
	}
	seedHash := evolution.Hash(vm.Seed(), [8]int{})
	for _, p := range w.Particles {
		if len(p.Code) > 0 && evolution.Hash(p.Code, [8]int{}) != seedHash {
			t.Fatal("unexpected genotype")
		}
	}
}

func TestNoInflowEventuallyDies(t *testing.T) {
	w := small(t, 1)
	w.Config.Inflow = 0
	// Accelerate exhaustion while preserving the accounting identity.
	for i := range w.Cells {
		w.Accounting.InitialEnergy -= int64(w.Cells[i].Energy)
		w.Cells[i].Energy = 0
	}
	advance(t, w, 500)
	if len(w.Particles) != 0 || w.Accounting.Deaths == 0 {
		t.Fatal("particles survive without available energy")
	}
}

func TestAllocateDoesNotCopyOrExecute(t *testing.T) {
	w := small(t, 2)
	p := w.Particles[1]
	p.Code = []vm.Instruction{{Op: vm.ALLOCATE, A: 1}}
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.RegisterGenome(p, "")
	w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
	Step(w)
	q := w.Particles[p.Target]
	if q == nil || len(q.Code) != 0 || q.Energy != 11 || w.Accounting.Copies != 0 {
		t.Fatal("ALLOCATE must create only blank matter")
	}
	if w.Accounting.Instructions != 1 {
		t.Fatal("new particle executed in its creation tick")
	}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestLimitsAndCorruptSnapshot(t *testing.T) {
	w := small(t, 1)
	w.Config.MaxEntities = 1
	advance(t, w, 200)
	if len(w.Particles) > 1 || w.Accounting.Allocations != 0 {
		t.Fatal("entity cap exceeded")
	}
	var buf bytes.Buffer
	if err := Save(&buf, w); err != nil {
		t.Fatal(err)
	}
	valid := append([]byte(nil), buf.Bytes()...)
	for _, bad := range [][]byte{[]byte(`{}`), append(valid, []byte(`{}`)...), bytes.Replace(valid, []byte(`"format":2`), []byte(`"format":99`), 1)} {
		if _, err := Load(bytes.NewReader(bad)); err == nil {
			t.Fatal("accepted bad snapshot")
		}
	}
	w.Cells[0].Energy = -1
	if err := Save(&buf, w); err == nil {
		t.Fatal("saved invalid resources")
	}
}

func TestCompetingAllocationsHaveOneWinner(t *testing.T) {
	w := small(t, 1)
	// Two intents compete for one cell; resolver cannot double-spend matter.
	p := w.Particles[1]
	e := rules.Event{Actor: p.ID, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.ALLOCATE, A: 1}}}
	rules.Resolve(w, e)
	first := p.Target
	rules.Resolve(w, e)
	if first == 0 || p.Target != 0 || w.Accounting.Allocations != 1 {
		t.Fatal("double allocation")
	}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestObserverIsReadOnly(t *testing.T) {
	w := small(t, 10)
	advance(t, w, 100)
	before := Hash(w)
	a, _ := json.Marshal(observer.Observe(w))
	b, _ := json.Marshal(observer.Observe(w))
	if before != Hash(w) || !bytes.Equal(a, b) {
		t.Fatal("observer changes state or order")
	}
}

func BenchmarkStep(b *testing.B) {
	w := small(b, 1)
	for i := 0; i < 200; i++ {
		Step(w)
	}
	var snapshot bytes.Buffer
	if err := Save(&snapshot, w); err != nil {
		b.Fatal(err)
	}
	initialParticles := len(w.Particles)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Measure the same populated 128-tick window, not an extinct world.
		if i%128 == 0 {
			b.StopTimer()
			var err error
			w, err = Load(bytes.NewReader(snapshot.Bytes()))
			if err != nil {
				b.Fatal(err)
			}
			b.StartTimer()
		}
		Step(w)
	}
	b.ReportMetric(float64(initialParticles), "initial-particles")
}
