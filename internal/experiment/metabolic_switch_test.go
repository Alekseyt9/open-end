package experiment

import (
	"open-end/internal/kernel"
	"open-end/internal/world"
	"testing"
)

func TestZeroSwitchCostMatchesLegacyAndSourceStaysImmutable(t *testing.T) {
	c := world.DefaultConfig()
	c.Width = 8
	c.Height = 8
	c.MaxEntities = 64
	c.Ecology = true
	c.CopyModel = "evolving"
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	h := kernel.Hash(w)
	_, _, zero, err := ContinueSwitch(w, "0", 400, 100, 10)
	if err != nil {
		t.Fatal(err)
	}
	_, _, old, err := ContinueCollectives(w, "intact", 400, 100, 10)
	if err != nil {
		t.Fatal(err)
	}
	if zero.FinalHash != old.FinalHash || zero.InitialHash != h || kernel.Hash(w) != h {
		t.Fatal("zero switch changed legacy physics")
	}
	_, _, paid, err := ContinueSwitch(w, "2", 400, 100, 10)
	if err != nil {
		t.Fatal(err)
	}
	if paid.InitialHash == h || paid.SourceHash != h || kernel.Hash(w) != h {
		t.Fatal("unmatched treatment")
	}
}
func TestSpecializationNeedsActivityPurityAndConsecutiveIdentity(t *testing.T) {
	for _, x := range []struct {
		u     [2]int64
		class string
	}{{[2]int64{31, 0}, "low-activity"}, {[2]int64{90, 10}, "reaction0"}, {[2]int64{11, 89}, "generalist"}, {[2]int64{0, 32}, "reaction1"}} {
		if profileClass(x.u) != x.class {
			t.Fatal("bad specialization", x)
		}
	}
	w := roleFixture(t)
	s := &switchSink{w: w, cells: map[uint64]*MetabolicProfile{}, previous: map[uint64]string{}}
	for j, units := range []int64{100, 100, 20, 100} {
		w.Tick += 5000
		s.cells[1] = &MetabolicProfile{ID: 1, Genome: w.Particles[1].Genome, Units: [2]int64{units, 0}}
		r := s.flush()
		want := 0
		if j == 1 {
			want = 1
		}
		if r.Persistent[0] != want {
			t.Fatal("persistence crossed activity gap", j, r)
		}
	}
}
