package observer

import (
	"open-end/internal/world"
	"testing"
)

func TestRatesAndGenomeFrequency(t *testing.T) {
	a := Metrics{Tick: 10, Copies: 20, Deaths: 10, Converted: [2]int64{100, 20}}
	b := Metrics{Tick: 30, Copies: 30, Deaths: 25, Converted: [2]int64{160, 40}}
	i := Since(a, b)
	if i.CopyRate != 0.5 || i.DeathRate != 0.75 || i.Converted != [2]int64{60, 20} {
		t.Fatal("incorrect interval rates")
	}
	if Since(a, a).Ticks != 0 {
		t.Fatal("zero-length interval")
	}
	w, err := world.New(world.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	m := Observe(w)
	if len(m.ActiveGenomes) != 1 || m.ActiveGenomes[0].Frequency != 1 || m.ActiveGenomes[0].Births != 1 || m.ActiveGenomes[0].Copies != 0 {
		t.Fatal("seed observation incorrectly counts offspring")
	}
}
