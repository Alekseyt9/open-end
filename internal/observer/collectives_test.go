package observer

import (
	"open-end/internal/evolution"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func collectiveFixture(t *testing.T) (*world.World, *Tracker, []*world.Particle) {
	t.Helper()
	config := world.DefaultConfig()
	config.Width = 8
	config.Height = 8
	config.MaxEntities = 64
	config.MutationPPM = 0
	w, err := world.New(config)
	if err != nil {
		t.Fatal(err)
	}
	first := w.Particles[1]
	w.Cells[first.Position].Occupant = 0
	w.Cells[first.Position].Matter++
	first.Position = 9
	w.Cells[9].Occupant = 1
	w.Cells[9].Matter--
	first.Code = []vm.Instruction{{Op: vm.NOP}}
	first.Origin = evolution.Hash(first.Code, first.InitialMemory)
	w.Origins[first.Origin] = &world.Origin{Hash: first.Origin}
	w.RegisterGenome(first, "")
	particles := []*world.Particle{first}
	for i, pos := range []int{10, 17, 18, 25} {
		p := &world.Particle{ID: w.NextID, Position: pos, Energy: 16}
		w.NextID++
		w.Cells[pos].Matter--
		w.Cells[pos].Energy -= 16
		w.Cells[pos].Occupant = p.ID
		w.Particles[p.ID] = p
		if i == 0 {
			p.Code = append([]vm.Instruction{}, first.Code...)
			p.Origin = first.Origin
			w.RegisterGenome(p, "")
		}
		particles = append(particles, p)
	}
	w.Relations[world.RelationKey(1, 2)] = world.Relation{A: 1, B: 2}
	tracker := NewTracker(w)
	if err := tracker.EnableCollectives(w, w, 2); err != nil {
		t.Fatal(err)
	}
	return w, tracker, particles
}
func groupAdvance(w *world.World, tr *Tracker, n int) {
	for i := 0; i < n; i++ {
		w.Tick++
		tr.TickCompleted(w.Tick)
	}
}
func groupCopy(w *world.World, tr *Tracker, source, target *world.Particle) {
	source.Target = target.ID
	rules.ResolveObserved(w, rules.Event{Actor: source.ID, Intent: vm.Intent{Instruction: vm.Instruction{Op: vm.COPY}}}, tr)
}

func TestDaughterCandidateRequiresMultipleFoundersAndParent(t *testing.T) {
	for _, scenario := range []string{"valid", "one-founder", "parent-gone"} {
		t.Run(scenario, func(t *testing.T) {
			w, tr, p := collectiveFixture(t)
			start := tr.Frame(w)
			groupAdvance(w, tr, 2)
			if len(tr.collectives.state.Cohorts) != 1 {
				t.Fatal("stable parent not tagged")
			}
			groupCopy(w, tr, p[0], p[2])
			source := p[1]
			if scenario == "one-founder" {
				source = p[0]
			}
			groupCopy(w, tr, source, p[3])
			w.Relations[world.RelationKey(3, 4)] = world.Relation{A: 3, B: 4}
			if scenario == "parent-gone" {
				delete(w.Relations, world.RelationKey(1, 2))
			}
			groupAdvance(w, tr, 3)
			frame := tr.Frame(w)
			if scenario != "valid" {
				if len(frame.Telemetry.Collectives.Candidates) != 0 {
					t.Fatal("false daughter candidate")
				}
				return
			}
			c := frame.Telemetry.Collectives
			if len(c.Candidates) != 1 || c.Productive != 0 || !c.Candidates[0].AllFounders || c.Candidates[0].Contributors != 2 {
				t.Fatalf("candidate evidence wrong: %+v", c.Candidates)
			}
			groupCopy(w, tr, p[2], p[4])
			groupAdvance(w, tr, 1)
			end := tr.Frame(w)
			if end.Telemetry.Collectives.Productive != 1 || end.Telemetry.Collectives.Candidates[0].Copies != 1 {
				t.Fatal("daughter reproduction not observed")
			}
			if c.Candidates[0].Copies != 0 || c.Cohorts[0].Activity[p[0].Genome].Copies != 2 {
				t.Fatal("frame aliases tracker")
			}
			if _, err := Summarize([]Metrics{start, frame, end}, 6); err != nil {
				t.Fatal(err)
			}
			end.Telemetry.Collectives.Candidates[0].Members[0] = p[0].ID
			if _, err := summarizeCollectives([]Metrics{frame, end}); err == nil {
				t.Fatal("parent member accepted in daughter")
			}
		})
	}
}

func TestChangingMembershipResetsMaturity(t *testing.T) {
	w, tr, _ := collectiveFixture(t)
	groupAdvance(w, tr, 1)
	delete(w.Relations, world.RelationKey(1, 2))
	groupAdvance(w, tr, 1)
	w.Relations[world.RelationKey(1, 2)] = world.Relation{A: 1, B: 2}
	groupAdvance(w, tr, 2)
	if len(tr.collectives.state.Cohorts) != 0 {
		t.Fatal("separate episodes accumulated age")
	}
	groupAdvance(w, tr, 1)
	if len(tr.collectives.state.Cohorts) != 1 {
		t.Fatal("mature group missed")
	}
}
