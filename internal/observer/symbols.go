package observer

import (
	"fmt"
	"open-end/internal/world"
	"sort"
)

type SymbolMetrics struct {
	Model string            `json:"model"`
	Cells int               `json:"cells"`
	Words int               `json:"words"`
	Pairs int               `json:"pairs"`
	State world.SymbolState `json:"state"`
}
type SymbolActivity struct {
	Genome string `json:"genome"`
	world.SymbolCounts
}
type SymbolWindow struct {
	Model string `json:"model"`
	Words int    `json:"words_at_end"`
	Pairs int    `json:"pairs_at_end"`
	world.SymbolCounts
	Activity []SymbolActivity `json:"activity"`
}

func observeSymbols(w *world.World) *SymbolMetrics {
	if w.Symbols == nil {
		return nil
	}
	r := &SymbolMetrics{Model: w.Config.Symbols, Cells: len(w.Cells), State: world.SymbolState{SymbolCounts: w.Symbols.SymbolCounts}}
	if w.Symbols.Actors != nil {
		r.State.Actors = map[string]*world.SymbolCounts{}
		for hash, a := range w.Symbols.Actors {
			copy := *a
			r.State.Actors[hash] = &copy
		}
	}
	for _, c := range w.Cells {
		if c.Word != nil && c.Word.Expires > w.Tick {
			r.Words++
			if len(c.Word.Tokens) == 2 {
				r.Pairs++
			}
		}
	}
	return r
}
func summarizeSymbols(frames []Metrics) (*SymbolWindow, error) {
	first, last := frames[0].Symbols, frames[len(frames)-1].Symbols
	if first == nil {
		for _, m := range frames {
			if m.Symbols != nil {
				return nil, fmt.Errorf("symbols appear within session")
			}
		}
		return nil, nil
	}
	previous := first
	for _, m := range frames {
		s := m.Symbols
		if s == nil || s.Model != first.Model || s.Cells != first.Cells || s.Cells < 4 || s.Cells > 1024*1024 || s.Words < 0 || s.Words > s.Cells || s.Pairs < 0 || s.Pairs > s.Words {
			return nil, fmt.Errorf("invalid symbol metrics")
		}
		if s.Model != "persistent" && s.Model != "scrambled" && s.Model != "unreadable" {
			return nil, fmt.Errorf("invalid symbol model")
		}
		w := &world.World{Config: world.Config{Symbols: s.Model}, Symbols: &s.State, Genomes: map[string]*world.GenomeRecord{}}
		for hash := range s.State.Actors {
			if len(hash) != 64 {
				return nil, fmt.Errorf("invalid symbol genome")
			}
			w.Genomes[hash] = &world.GenomeRecord{}
		}
		if err := w.ValidateSymbols(); err != nil {
			return nil, err
		}
		if _, err := s.State.SymbolCounts.Sub(previous.State.SymbolCounts); err != nil {
			return nil, err
		}
		for hash, a := range previous.State.Actors {
			b := s.State.Actors[hash]
			if b == nil {
				return nil, fmt.Errorf("symbol actor disappeared")
			}
			if _, err := b.Sub(*a); err != nil {
				return nil, err
			}
		}
		previous = s
	}
	delta, err := last.State.SymbolCounts.Sub(first.State.SymbolCounts)
	if err != nil {
		return nil, err
	}
	r := &SymbolWindow{Model: first.Model, Words: last.Words, Pairs: last.Pairs, SymbolCounts: delta, Activity: []SymbolActivity{}}
	for hash, b := range last.State.Actors {
		a := first.State.Actors[hash]
		if a == nil {
			a = &world.SymbolCounts{}
		}
		d, err := b.Sub(*a)
		if err != nil {
			return nil, err
		}
		if d != (world.SymbolCounts{}) {
			r.Activity = append(r.Activity, SymbolActivity{hash, d})
		}
	}
	sort.Slice(r.Activity, func(i, j int) bool { return r.Activity[i].Genome < r.Activity[j].Genome })
	return r, nil
}
