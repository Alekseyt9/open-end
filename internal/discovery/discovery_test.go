package discovery

import (
	"open-end/internal/evolution"
	"open-end/internal/experiment"
	"open-end/internal/kernel"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"reflect"
	"slices"
	"testing"
)

func fixture(t *testing.T) *world.World {
	t.Helper()
	c := world.DefaultConfig()
	c.Width, c.Height, c.MaxEntities = 8, 8, 64
	c.MutationPPM = 0
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	p := w.Particles[1]
	p.Code = []vm.Instruction{{Op: vm.NOP}}
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
	w.RegisterGenome(p, "")
	for _, pos := range []int{9, 10, 17} {
		q := &world.Particle{ID: w.NextID, Position: pos, Energy: 16, Code: slices.Clone(p.Code), Origin: p.Origin}
		w.NextID++
		w.Cells[pos].Matter--
		w.Cells[pos].Energy -= 16
		w.Cells[pos].Occupant = q.ID
		w.Particles[q.ID] = q
		w.RegisterGenome(q, "")
	}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
	return w
}
func event(r *recorder, kind string, a, b uint64, energy int64) {
	r.Interaction(rules.Interaction{Tick: r.w.Tick + 1, Kind: kind, SourceID: a, TargetID: b, Source: r.w.Particles[a].Genome, Target: r.w.Particles[b].Genome, Energy: energy})
}
func advance(r *recorder) { r.w.Tick++; r.TickCompleted(r.w.Tick) }
func TestReciprocalGraphExcludesOneWayTakingAndDeadEndpoints(t *testing.T) {
	w := fixture(t)
	e := map[pair]edge{{1, 2}: {transfer: 4}, {2, 1}: {transfer: 3}, {2, 3}: {transfer: 2}, {3, 2}: {taken: 9}}
	if got := reciprocal(w, e); !reflect.DeepEqual(got, [][]uint64{{1, 2}}) {
		t.Fatal(got)
	}
	e[pair{3, 1}] = edge{transfer: 1}
	if got := reciprocal(w, e); !reflect.DeepEqual(got, [][]uint64{{1, 2, 3}}) {
		t.Fatal("directed cycle missed", got)
	}
	// An external return path must not validate a frozen two-member candidate.
	delete(e, pair{2, 1})
	if reciprocalWithin(w, e, []uint64{1, 2}) {
		t.Fatal("external return path counted as internal")
	}
	delete(w.Particles, 3)
	if len(reciprocal(w, e)) != 0 {
		t.Fatal("dead endpoint created reciprocity")
	}
}
func TestInclusionHierarchyKeepsOverlapAndRemovesTransitiveEdges(t *testing.T) {
	nodes := []Node{}
	for _, ids := range [][]uint64{{1, 2}, {2, 3}, {1, 2, 3}, {1, 2, 3, 4}} {
		nodes = append(nodes, Node{ID: nodeID(ids), Members: ids, Children: []string{}, Level: 1})
	}
	hierarchy(nodes)
	if nodes[0].Overlap != 1 || nodes[1].Overlap != 1 {
		t.Fatal("overlap forced into containment")
	}
	if len(nodes[2].Children) != 2 || len(nodes[3].Children) != 1 || nodes[3].Children[0] != nodes[2].ID || nodes[3].Level != 3 {
		t.Fatal("invalid transitive reduction", nodes)
	}
}
func TestProspectiveFlowValidationAndSamplingEpisodes(t *testing.T) {
	w := fixture(t)
	r, err := newRecorder(w, w, Config{Every: 1, MinAge: 1, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	initial, _ := r.sample(0, nil)
	event(r, "transfer", 1, 2, 4)
	event(r, "transfer", 2, 1, 3)
	event(r, "transfer", 3, 1, 2)
	event(r, "take", 1, 2, 9)
	advance(r)
	first, _ := r.sample(0, &initial)
	if len(first.Nodes) != 1 {
		t.Fatal("missing reciprocal candidate")
	}
	n := first.Nodes[0]
	if n.Current.Internal != 7 || n.Current.Incoming != 2 || n.Current.TakenInternal != 9 || n.Current.Retention == nil || n.PersistentFlow || n.Next != nil {
		t.Fatal("flow attribution wrong", n)
	}
	event(r, "transfer", 1, 2, 5)
	event(r, "transfer", 2, 1, 2)
	advance(r)
	second, _ := r.sample(1, &first)
	if !first.Nodes[0].NextReciprocal || first.Nodes[0].Next.Internal != 7 || first.Nodes[0].NextAlive != 2 || !second.Nodes[0].PersistentFlow {
		t.Fatal("next interval or repeated boundary missing")
	}
	// No exchange in the third interval: older successful evidence stays unchanged.
	advance(r)
	third, _ := r.sample(2, &second)
	if second.Nodes[0].Next.Internal != 0 || second.Nodes[0].NextReciprocal || first.Nodes[0].Next.Internal != 7 {
		t.Fatal("future evidence leaked between intervals")
	}
	event(r, "transfer", 1, 2, 5)
	event(r, "transfer", 2, 1, 2)
	advance(r)
	fourth, _ := r.sample(3, &third)
	if fourth.Nodes[0].FlowSamples != 1 || fourth.Nodes[0].PersistentFlow {
		t.Fatal("nonconsecutive intervals accumulated persistence")
	}
}
func TestTruncatedIntervalsCannotSupportFlowBoundaries(t *testing.T) {
	w := fixture(t)
	r, err := newRecorder(w, w, Config{Every: 1, MinAge: 1, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	event(r, "transfer", 1, 2, 3)
	event(r, "transfer", 2, 1, 3)
	advance(r)
	f, _ := r.sample(0, nil)
	if f.Complete || f.Dropped != 1 || len(f.Nodes) != 0 || r.measure([]uint64{1, 2}) == nil {
		t.Fatal("truncation or reset handling wrong")
	}
}
func TestReplayMatchesAssayAndPreservesSource(t *testing.T) {
	c := world.DefaultConfig()
	c.Width, c.Height, c.MaxEntities = 8, 8, 64
	c.Ecology, c.Environment, c.CopyModel = true, "coupled", "evolving"
	c.MutationPPM = 100000
	w, err := world.New(c)
	if err != nil {
		t.Fatal(err)
	}
	sourceHash := kernel.Hash(w)
	for _, mode := range []string{"intact", "bonds", "sharing", "signal-reading"} {
		_, _, trial, err := experiment.ContinueCollectives(w, mode, 500, 100, 10)
		if err != nil {
			t.Fatal(err)
		}
		a, err := Replay(w, mode, 500, Config{Every: 100, MinAge: 10, Limit: 65536})
		if err != nil {
			t.Fatal(err)
		}
		b, err := Replay(w, mode, 500, Config{Every: 50, MinAge: 20, Limit: 1})
		if err != nil {
			t.Fatal(err)
		}
		if a.FinalHash != trial.FinalHash || b.FinalHash != trial.FinalHash || a.InitialHash != trial.InitialHash || kernel.Hash(w) != sourceHash {
			t.Fatal("observer or interval changed physics")
		}
		last := a.Frames[len(a.Frames)-1]
		for _, n := range last.Nodes {
			if n.Next != nil || n.NextComplete {
				t.Fatal("invented final holdout")
			}
		}
	}
}

func TestCopyAncestryUsesSuccessfulActorAndCleansDeadMembers(t *testing.T) {
	w := fixture(t)
	child := w.Particles[3]
	child.Code = nil
	child.Genome = ""
	child.Origin = ""
	child.Parent = 1
	r, err := newRecorder(w, w, Config{Every: 1, MinAge: 1, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	w.Particles[2].Target = 3
	rules.ResolveObserved(w, rules.Event{Actor: 2, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.COPY}}}, r)
	advance(r)
	f, _ := r.sample(0, nil)
	if len(f.Nodes) != 1 || f.Nodes[0].CopyRoot != 2 || !slices.Equal(f.Nodes[0].Members, []uint64{2, 3}) || !slices.Equal(f.MicroIDs, []uint64{1, 2, 3, 4}) {
		t.Fatal("allocation parent substituted for actual copy actor", f)
	}
	delete(w.Particles, 3)
	r.Death(rules.Death{ID: 3, Tick: w.Tick + 1, Genome: child.Genome})
	advance(r)
	end, _ := r.sample(1, &f)
	if len(end.Nodes) != 0 || f.Nodes[0].NextAlive != 1 {
		t.Fatal("dead member retained in ancestry boundary")
	}
}
