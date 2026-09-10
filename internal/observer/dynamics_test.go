package observer

import (
	"encoding/json"
	"fmt"
	"open-end/internal/dsl"
	"strings"
	"testing"
)

// A balanced synthetic recording isolates detector decisions from evolution.
// Optional transient copies die inside each interval and never survive a frame.
func dynamicsFixture(churn bool) []Metrics {
	frames := make([]Metrics, 21)
	for i := range frames {
		m := Metrics{Tick: uint64(i * 1000), Entities: 10, Executable: 10, Genomes: 1, Lineages: 1,
			ActiveGenomes: []Genome{{Hash: "resident", Count: 10, Frequency: 1}},
			Telemetry: &Telemetry{Version: 1, SessionHash: strings.Repeat("a", 64), Complete: true,
				Diversity:   Diversity{EffectiveGenomes: 1, DominantShare: 1},
				Structures:  Structures{Components: 10, Largest: 1, Sizes: []SizeCount{{Size: 1, Count: 10}}},
				ActiveRules: RuleIdentity{Version: "ecology-2", Hash: "builtin"}}}
		if i > 0 {
			m.Telemetry.FromTick = uint64((i - 1) * 1000)
		}
		if churn && i > 0 {
			hash := fmt.Sprintf("transient-%d", i)
			m.Copies, m.Deaths = uint64(i), uint64(i)
			m.Telemetry.Interactions = []Edge{{Kind: "copy", Source: "resident", Target: hash, Count: 1}}
			m.Telemetry.DeathsByGenome = []GenomeDeaths{{Hash: hash, Count: 1}}
			m.Telemetry.Lifetimes = Lifetimes{Count: 1, Sum: 1, Min: 1, Max: 1, Histogram: []LifeBucket{{Min: 1, Max: 1, Count: 1}}}
			m.Telemetry.Discoveries = []Discovery{{Hash: hash, Parent: "resident", FirstTick: uint64((i - 1) * 1000)}}
		}
		frames[i] = m
	}
	return frames
}

func detectFixture(t *testing.T, frames []Metrics, want string) Dynamics {
	t.Helper()
	r, err := DetectDynamics(frames, 20000, DefaultDynamicsConfig())
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != want {
		t.Fatalf("want %s, got %+v", want, r)
	}
	return r
}

func TestDynamicsStableAndNeutralChurn(t *testing.T) {
	for _, churn := range []bool{false, true} {
		frames := dynamicsFixture(churn)
		r := detectFixture(t, frames, "stagnating")
		if !r.Monoculture || r.RepeatedFraction != 1 || r.PersistentNewBehavior {
			t.Fatalf("bad plateau: %+v", r)
		}
		if churn && r.NewGenomes != 20 {
			t.Fatal("lost transient genome churn")
		}
		before, _ := json.Marshal(frames)
		again := detectFixture(t, frames, "stagnating")
		a, _ := json.Marshal(r)
		b, _ := json.Marshal(again)
		after, _ := json.Marshal(frames)
		if string(a) != string(b) || string(before) != string(after) {
			t.Fatal("nondeterminism or input mutation")
		}
	}
}

func TestDynamicsPersistentBehaviorVersusTransientNoise(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		frames := dynamicsFixture(false)
		var absorbed int64
		for i := 1; i < len(frames); i++ {
			if i > 10 && (persistent || i <= 15) {
				absorbed += 10000
				frames[i].Telemetry.Flows.Absorbed = 10000
			}
			frames[i].Absorbed = absorbed
		}
		want := "mixed"
		if persistent {
			want = "developing"
		}
		r := detectFixture(t, frames, want)
		if r.PersistentNewBehavior != persistent {
			t.Fatal("did not distinguish transient change")
		}
	}
}

func TestDynamicsQuantizationBoundaryIsNotNovelty(t *testing.T) {
	frames := dynamicsFixture(false)
	var absorbed int64
	for i := 1; i < len(frames); i++ {
		flow := int64(136)
		if i > 10 {
			flow = 137
		}
		absorbed += flow
		frames[i].Absorbed = absorbed
		frames[i].Telemetry.Flows.Absorbed = flow
	}
	r := detectFixture(t, frames, "mixed")
	if r.Behavior[0].Hash == r.Behavior[3].Hash || r.PersistentNewBehavior {
		t.Fatal("boundary fixture must change hash without meaningful novelty")
	}
}

func TestDynamicsStructureGrowthAndDistribution(t *testing.T) {
	frames := dynamicsFixture(false)
	for i := 11; i < len(frames); i++ {
		frames[i].Telemetry.Structures = Structures{Components: 1, LinkedComponents: 1, LinkedParticles: 10, Largest: 10, Sizes: []SizeCount{{Size: 10, Count: 1}}}
	}
	r := detectFixture(t, frames, "developing")
	if !r.PersistentStructureGrowth || r.StructuralPlateau || r.StructureDistance != 1 {
		t.Fatal("missed changed structure")
	}
	// Largest component alone is not sufficient to describe a distribution.
	frames = dynamicsFixture(false)
	for i := range frames {
		frames[i].Telemetry.Structures = Structures{Largest: 4, Sizes: []SizeCount{{Size: 4, Count: 1}, {Size: 1, Count: 6}}}
		if i > 10 {
			frames[i].Telemetry.Structures.Sizes = []SizeCount{{Size: 4, Count: 2}, {Size: 2, Count: 1}}
		}
	}
	r = detectFixture(t, frames, "mixed")
	if r.StructuralPlateau {
		t.Fatal("ignored distribution change at constant maximum size")
	}
}

func TestDynamicsGuards(t *testing.T) {
	detectFixture(t, dynamicsFixture(false)[:4], "insufficient_history")
	frames := dynamicsFixture(false)
	// Plenty of samples, but the first quarter is entirely unobserved.
	for i := 1; i < len(frames); i++ {
		frames[i].Tick += 19000
		frames[i].Telemetry.FromTick = frames[i-1].Tick
	}
	if r, err := DetectDynamics(frames, 50000, DefaultDynamicsConfig()); err != nil || r.Status != "insufficient_history" {
		t.Fatalf("sparse cadence: %v %+v", err, r)
	}
	frames = dynamicsFixture(false)
	frames[12].Telemetry.RuleEvents = []dsl.Event{{Tick: 11500, Action: "activate", Hash: "new"}, {Tick: 11900, Action: "rollback", Hash: "builtin"}}
	detectFixture(t, frames, "rule_change")
	frames = dynamicsFixture(false)
	for i := range frames {
		frames[i].Executable = 0
		frames[i].Genomes = 0
		frames[i].ActiveGenomes = nil
	}
	detectFixture(t, frames, "extinct")
	frames = dynamicsFixture(false)
	frames[12].Telemetry.Complete = false
	if _, err := DetectDynamics(frames, 20000, DefaultDynamicsConfig()); err == nil {
		t.Fatal("accepted gap")
	}
	frames = dynamicsFixture(false)
	frames[12].Telemetry.SessionHash = strings.Repeat("b", 64)
	if _, err := DetectDynamics(frames, 20000, DefaultDynamicsConfig()); err == nil {
		t.Fatal("accepted mixed session")
	}
	frames[12].Telemetry = nil
	if _, err := DetectDynamics(frames, 20000, DefaultDynamicsConfig()); err == nil {
		t.Fatal("accepted legacy frame")
	}
	cfg := DefaultDynamicsConfig()
	cfg.BehaviorResolution = 0
	if _, err := DetectDynamics(dynamicsFixture(false), 20000, cfg); err == nil {
		t.Fatal("accepted invalid config")
	}
}

func TestDynamicsHonorsRequestedWindowAndCadence(t *testing.T) {
	frames := dynamicsFixture(false)
	frames[1].Telemetry.RuleEvents = []dsl.Event{{Tick: 500, Action: "activate"}}
	r, err := DetectDynamics(frames, 10000, DefaultDynamicsConfig())
	if err != nil || r.FromTick != 10000 || r.Status != "stagnating" {
		t.Fatalf("old rule change leaked into window: %v %+v", err, r)
	}
	// Equal rates over unequal block durations should have the same signature.
	a := behaviorSignature([]Metrics{{Tick: 0, Entities: 10}, {Tick: 1000, Entities: 10}}, WindowSummary{FromTick: 0, ToTick: 1000, Flows: Flows{Absorbed: 100}}, 4)
	b := behaviorSignature([]Metrics{{Tick: 0, Entities: 10}, {Tick: 3000, Entities: 10}}, WindowSummary{FromTick: 0, ToTick: 3000, Flows: Flows{Absorbed: 300}}, 4)
	if a.Hash != b.Hash {
		t.Fatal("hash depends on elapsed duration at identical rate")
	}
}
