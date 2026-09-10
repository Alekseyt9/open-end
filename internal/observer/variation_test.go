package observer

import (
	"open-end/internal/evolution"
	"strings"
	"testing"
)

func TestVariationPersistenceCountsDecodedPolicies(t *testing.T) {
	makeFrame := func(tick uint64, copies uint64) Metrics {
		m := Metrics{Tick: tick, CopyModel: "evolving", Copies: copies * 3}
		for i := 0; i < 3; i++ {
			hash := strings.Repeat(string(rune('a'+i)), 64)
			ppm := 10000
			if i == 2 {
				ppm = 0
			}
			m.ActiveGenomes = append(m.ActiveGenomes, Genome{Hash: hash, Count: 5})
			m.Variation = append(m.Variation, evolution.CopyRecord{Genome: hash, Policy: evolution.CopyPolicy{Kind: "code", PPM: ppm}, Copies: copies, FirstTick: 1, LastTick: tick - 1})
		}
		return m
	}
	frames := []Metrics{makeFrame(100, 10), makeFrame(200, 20), makeFrame(300, 30)}
	v, err := summarizeVariation(frames)
	if err != nil {
		t.Fatal(err)
	}
	if v.CodePolicies != 2 || v.PersistentCodePolicies != 2 || v.CodeCopies != 60 || len(v.Activity) != 3 {
		t.Fatalf("%+v", v)
	}
	// A sampled dip invalidates persistence even if the population recovers.
	frames[1].ActiveGenomes[2].Count = 4
	v, err = summarizeVariation(frames)
	if err != nil || v.PersistentCodePolicies != 1 {
		t.Fatalf("sampled dip ignored: %v %+v", err, v)
	}
	frames[1].ActiveGenomes[2].Count = 5
	// Existing population without copy activity is not an expressed strategy.
	frames[1].Variation[2].Copies = 10
	frames[1].Copies -= 10
	v, err = summarizeVariation(frames)
	if err != nil || v.PersistentCodePolicies != 1 {
		t.Fatalf("silent interval ignored: %v %+v", err, v)
	}
	// A deleted historical record must not masquerade as a new observation.
	frames[1].Variation = frames[1].Variation[:2]
	frames[1].Copies -= 10
	if _, err := summarizeVariation(frames); err == nil {
		t.Fatal("missing historical record accepted")
	}
}

func TestVariationRejectsDamagedCountersAndModes(t *testing.T) {
	base := Metrics{Tick: 100, CopyModel: "fixed", Copies: 10, Variation: []evolution.CopyRecord{{Genome: strings.Repeat("a", 64), Policy: evolution.CopyPolicy{Kind: "code", PPM: 10000}, Copies: 10, FirstTick: 1, LastTick: 99}}}
	for _, mutate := range []func(*Metrics){
		func(m *Metrics) { m.CopyModel = "evolving" },
		func(m *Metrics) { m.Copies = 11 },
		func(m *Metrics) { m.Variation[0].Changed = 11 },
		func(m *Metrics) { m.Variation[0].FirstTick = 2 },
		func(m *Metrics) { m.Variation[0].Recombined = 1 },
		func(m *Metrics) { m.Variation = append(m.Variation, m.Variation[0]); m.Copies = 20 },
	} {
		next := base
		next.Tick = 200
		next.Variation = append([]evolution.CopyRecord(nil), base.Variation...)
		mutate(&next)
		if _, err := summarizeVariation([]Metrics{base, next}); err == nil {
			t.Fatal("damaged variation accepted")
		}
	}
}
