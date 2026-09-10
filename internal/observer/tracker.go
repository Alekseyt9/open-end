package observer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/bits"
	"open-end/internal/dsl"
	"open-end/internal/rules"
	"open-end/internal/world"
	"sort"
)

type edgeKey struct{ kind, source, target string }

// Tracker is an external interval accumulator. A new tracker starts a new
// observation session, including when a physical snapshot is resumed.
type Tracker struct {
	sessionHash       string
	start, from, last uint64
	complete          bool
	previous          world.Accounting
	dslUsage          map[string]uint64
	ruleEvents        int
	edges             map[edgeKey]Edge
	deaths            map[string]uint64
	discoveries       map[string]Discovery
	life              Lifetimes
	bins              [65]uint64
}

func NewTracker(w *world.World) *Tracker {
	t := &Tracker{start: w.Tick, from: w.Tick, last: w.Tick, previous: w.Accounting}
	h := sha256.New()
	_ = json.NewEncoder(h).Encode(w)
	t.sessionHash = hex.EncodeToString(h.Sum(nil))
	t.reset()
	t.dslUsage = map[string]uint64{}
	if w.RuleState != nil {
		for k, v := range w.RuleState.Usage {
			t.dslUsage[k] = v
		}
		t.ruleEvents = len(w.RuleState.Events)
	}
	return t
}
func (t *Tracker) reset() {
	t.complete = true
	t.edges = map[edgeKey]Edge{}
	t.deaths = map[string]uint64{}
	t.discoveries = map[string]Discovery{}
	t.life = Lifetimes{Histogram: []LifeBucket{}}
	t.bins = [65]uint64{}
}
func (t *Tracker) TickCompleted(tick uint64) {
	if tick != t.last+1 {
		t.complete = false
	}
	t.last = tick
}
func (t *Tracker) Interaction(e rules.Interaction) {
	if e.Tick != t.last+1 {
		t.complete = false
	}
	key := edgeKey{e.Kind, e.Source, e.Target}
	v := t.edges[key]
	v.Kind = e.Kind
	v.Source = e.Source
	v.Target = e.Target
	v.Count++
	v.Energy += e.Energy
	t.edges[key] = v
	if e.NewGenome {
		t.discoveries[e.Target] = Discovery{e.Target, e.Source, e.Tick - 1}
	}
}
func (t *Tracker) Death(d rules.Death) {
	if d.Tick != t.last+1 || d.Created > d.Tick {
		t.complete = false
		return
	}
	age := d.Tick - d.Created
	if t.life.Count == 0 {
		t.life.Min = age
	}
	t.life.Count++
	t.life.Sum += age
	t.life.Min = min(t.life.Min, age)
	t.life.Max = max(t.life.Max, age)
	t.bins[bits.Len64(age)]++
	t.deaths[d.Genome]++
}

// Frame closes an interval and resets only observer accumulators. It never
// changes world state, archives, instruction scheduling or either RNG stream.
func (t *Tracker) Frame(w *world.World) Metrics {
	m := Observe(w)
	x := &Telemetry{Version: 1, SessionStart: t.start, FromTick: t.from, Complete: t.complete && t.last == w.Tick,
		Interactions: []Edge{}, DeathsByGenome: []GenomeDeaths{}, Discoveries: []Discovery{}, RuleEvents: []dsl.Event{}}
	x.Seed = w.Config.Seed
	x.SessionHash = t.sessionHash
	instantaneous(w, m, x)
	a, b := t.previous, w.Accounting
	x.Flows = Flows{Injected: b.Injected - a.Injected, Dissipated: b.Dissipated - a.Dissipated, Absorbed: b.Absorbed - a.Absorbed, Transferred: b.Transferred - a.Transferred, Taken: b.Taken - a.Taken, Allocated: int64(b.Allocations-a.Allocations) * 12, Charged: b.Charged - a.Charged, DSL: map[string]uint64{}}
	for i := range x.Flows.Converted {
		x.Flows.Converted[i] = b.Converted[i] - a.Converted[i]
	}
	if w.RuleState != nil {
		for k, v := range w.RuleState.Usage {
			old := t.dslUsage[k]
			if v < old {
				x.Complete = false
			} else if v > old {
				x.Flows.DSL[k] = v - old
			}
			t.dslUsage[k] = v
		}
		if t.ruleEvents > len(w.RuleState.Events) {
			x.Complete = false
		} else {
			x.RuleEvents = append(x.RuleEvents, w.RuleState.Events[t.ruleEvents:]...)
		}
		t.ruleEvents = len(w.RuleState.Events)
	}
	x.Lifetimes = t.life
	if t.life.Count > 0 {
		x.Lifetimes.Mean = float64(t.life.Sum) / float64(t.life.Count)
	}
	for i, count := range t.bins {
		if count > 0 {
			lo, hi := uint64(0), uint64(0)
			if i > 0 {
				lo = uint64(1) << (i - 1)
				hi = (uint64(1) << i) - 1
			}
			x.Lifetimes.Histogram = append(x.Lifetimes.Histogram, LifeBucket{lo, hi, count})
		}
	}
	for hash, count := range t.deaths {
		x.DeathsByGenome = append(x.DeathsByGenome, GenomeDeaths{hash, count})
	}
	sort.Slice(x.DeathsByGenome, func(i, j int) bool { return x.DeathsByGenome[i].Hash < x.DeathsByGenome[j].Hash })
	for _, e := range t.edges {
		x.Interactions = append(x.Interactions, e)
	}
	sortEdges(x.Interactions)
	for _, d := range t.discoveries {
		x.Discoveries = append(x.Discoveries, d)
	}
	sort.Slice(x.Discoveries, func(i, j int) bool { return x.Discoveries[i].Hash < x.Discoveries[j].Hash })
	if b.Deaths < a.Deaths || b.Deaths-a.Deaths != t.life.Count {
		x.Complete = false
	}
	m.Telemetry = x
	t.previous = b
	t.from = w.Tick
	t.last = w.Tick
	t.reset()
	return m
}
