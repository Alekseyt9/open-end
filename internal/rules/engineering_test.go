package rules

import (
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func TestConstructionSignalConservationAndErosion(t *testing.T) {
	w := fixture(t)
	w.Config.Environment = "coupled"
	w.Environment = &world.EnvironmentState{}
	p := w.Particles[1]
	setProgram(w, p, []vm.Instruction{{Op: vm.BUILD, A: 1, B: 8}, {Op: vm.EMIT, A: 1, B: 12}})
	source := p.Position
	w.Cells[source].Matter++
	w.Cells[w.Neighbor(source, 2)].Matter--
	other := w.Neighbor(source, 1)
	beforeEnergy, beforeMatter := w.Totals()
	build(w, p, 1, 8)
	if w.Cells[other].Terrain != 1 || w.Cells[source].Matter != 0 {
		t.Fatal("construction created matter")
	}
	emitSignal(w, p, 1, 12)
	energy, matter := w.Totals()
	if energy != beforeEnergy || matter != beforeMatter {
		t.Fatal("construction/emission lost resources")
	}
	check(t, w)
	environmentStep(w)
	if w.Cells[other].Terrain != 0 || w.Cells[other].Matter != 2 || w.Environment.Eroded != 1 || w.Environment.Decayed != 1 {
		t.Fatal("local erosion/decay failed")
	}
	check(t, w)
	// Reclaiming returns matter to the actor's cell, not the target cell.
	build(w, p, 1, -8)
	w.Cells[source].Matter++
	w.Accounting.InitialMatter++
	build(w, p, 1, 1)
	build(w, p, 1, -int(^uint(0)>>1)-1)
	if w.Cells[other].Terrain != 0 || w.Environment.Reclaimed != 1 {
		t.Fatal("minimum operand reclamation failed")
	}
	check(t, w)
}

func TestTerrainFeedbackAndTickStartSensing(t *testing.T) {
	for _, mode := range []string{"coupled", "inert"} {
		w := fixture(t)
		w.Config.Environment = mode
		w.Environment = &world.EnvironmentState{}
		p := w.Particles[1]
		w.Cells[p.Position].Matter++
		w.Cells[w.Neighbor(p.Position, 2)].Matter--
		w.Cells[p.Position].Energy -= 8
		w.Accounting.Dissipated += 8
		setProgram(w, p, []vm.Instruction{{Op: vm.BUILD, A: -1, B: 1}, {Op: vm.SENSE, A: 6, B: 3}})
		build(w, p, -1, 1)
		w.Tick = 1
		p.IP = 1
		event := Evaluate(w, p)
		build(w, p, -1, -1)
		Resolve(w, event)
		if p.Memory[3] != 1 {
			t.Fatal("sensing did not use tick-start view")
		}
		build(w, p, -1, 1)
		a, b := 8, 0
		terrainMix(w, &a, &b, p.Position, w.Neighbor(p.Position, 1))
		want := 4
		if mode == "coupled" {
			want = 2
		}
		if b != want || a+b != 8 {
			t.Fatal("terrain transfer feedback failed")
		}
		before := w.Accounting.Injected
		Inflow(w)
		if mode == "coupled" && w.Environment.BlockedLight == 0 {
			t.Fatal("terrain did not shade inflow")
		}
		if mode == "inert" && (w.Environment.BlockedLight != 0 || w.Environment.AttenuatedTransfer != 0) {
			t.Fatal("inert terrain changed transport")
		}
		if w.Accounting.Injected <= before {
			t.Fatal("inflow vanished")
		}
		check(t, w)
	}
}
