package observer_test

import (
	"open-end/internal/evolution"
	"open-end/internal/observer"
	"open-end/internal/vm"
	"open-end/internal/world"
	"testing"
)

func TestSignalEnergyReconcilesWithEventTelemetry(t *testing.T) {
	w := ecology(t)
	w.Config.Environment = "coupled"
	w.Environment = &world.EnvironmentState{}
	p := w.Particles[1]
	p.Code = []vm.Instruction{{Op: vm.CONVERT, A: 0, B: 8}, {Op: vm.EMIT, A: -1, B: 8}, {Op: vm.JUMP, A: 0}}
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
	w.RegisterGenome(p, "")
	frames := capture(t, w, 100, 10)
	emitted := false
	for _, m := range frames {
		if m.Telemetry.Pools.Signal > 0 {
			emitted = true
		}
		if m.Telemetry.Pools.Signal != m.Environment.Signal {
			t.Fatal("signal energy pool mismatch")
		}
	}
	if !emitted {
		t.Fatal("fixture emitted no signal")
	}
	r, err := observer.Summarize(frames, 100)
	if err != nil {
		t.Fatal(err)
	}
	if r.Environment.Emitted == 0 || r.Environment.Decayed == 0 {
		t.Fatal("missing energy flow")
	}
}
