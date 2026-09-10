// Package observer reads state without influencing selection or execution.
package observer

import (
	"open-end/internal/world"
	"sort"
)

type Lineage struct {
	Hash  string `json:"hash"`
	Count int    `json:"count"`
}

type Metrics struct {
	Tick                uint64    `json:"tick"`
	Entities            int       `json:"entities"`
	Executable          int       `json:"executable"`
	Genomes             int       `json:"genomes"`
	Lineages            int       `json:"lineages"`
	OriginsSeen         int       `json:"origins_seen"`
	Energy              int64     `json:"energy"`
	Matter              int64     `json:"matter"`
	Allocations         uint64    `json:"allocations"`
	Copies              uint64    `json:"copies"`
	Deaths              uint64    `json:"deaths"`
	Absorbed            int64     `json:"absorbed"`
	Transferred         int64     `json:"transferred"`
	MeanAge             float64   `json:"mean_age_ticks"`
	ActiveLineages      []Lineage `json:"active_lineages"`
	ActiveGenomes       []Genome  `json:"active_genomes"`
	Converted           [2]int64  `json:"converted"`
	Relations           int       `json:"relations"`
	Chemicals           [3]int64  `json:"chemicals"`
	LitMatter           int64     `json:"lit_matter"`
	DarkMatter          int64     `json:"dark_matter"`
	LitEmptyMatterCells int       `json:"lit_empty_matter_cells"`
	FailedMatter        uint64    `json:"failed_matter"`
	FailedSpace         uint64    `json:"failed_space"`
	FailedReserve       uint64    `json:"failed_reserve"`
	FailedLimit         uint64    `json:"failed_limit"`
	FailedAbsorb        uint64    `json:"failed_absorb"`
	FailedReaction      uint64    `json:"failed_reaction"`
	InstructionStarved  uint64    `json:"instruction_starved"`
	Interval            Interval  `json:"interval"`
}

type Genome struct {
	Hash          string   `json:"hash"`
	Count         int      `json:"count"`
	Frequency     float64  `json:"frequency"`
	MaxGeneration uint64   `json:"max_generation"`
	FirstTick     uint64   `json:"first_tick"`
	Births        uint64   `json:"births"`
	Copies        uint64   `json:"copies"`
	LastCopy      uint64   `json:"last_copy_tick"`
	Instructions  uint64   `json:"instructions"`
	Moved         uint64   `json:"moved"`
	Absorbed      int64    `json:"absorbed"`
	Converted     [2]int64 `json:"converted"`
	Taken         int64    `json:"taken"`
	Transferred   int64    `json:"transferred"`
	Binds         uint64   `json:"binds"`
}

type Interval struct {
	Ticks     uint64   `json:"ticks"`
	Copies    uint64   `json:"copies"`
	Deaths    uint64   `json:"deaths"`
	CopyRate  float64  `json:"copies_per_tick"`
	DeathRate float64  `json:"deaths_per_tick"`
	Converted [2]int64 `json:"converted"`
}

func Since(previous, current Metrics) Interval {
	if current.Tick <= previous.Tick {
		return Interval{}
	}
	n := current.Tick - previous.Tick
	r := Interval{Ticks: n, Copies: current.Copies - previous.Copies, Deaths: current.Deaths - previous.Deaths}
	r.CopyRate = float64(r.Copies) / float64(n)
	r.DeathRate = float64(r.Deaths) / float64(n)
	for i := range r.Converted {
		r.Converted[i] = current.Converted[i] - previous.Converted[i]
	}
	return r
}

func Observe(w *world.World) Metrics {
	e, m := w.Totals()
	a := w.Accounting
	r := Metrics{Tick: w.Tick, Entities: len(w.Particles), OriginsSeen: len(w.Origins), Energy: e, Matter: m,
		Allocations: a.Allocations, Copies: a.Copies, Deaths: a.Deaths, Absorbed: a.Absorbed, Transferred: a.Transferred,
		ActiveLineages: make([]Lineage, 0)}
	r.Converted = a.Converted
	r.Relations = len(w.Relations)
	r.FailedMatter = a.FailedMatter
	r.FailedSpace = a.FailedSpace
	r.FailedReserve = a.FailedReserve
	r.FailedLimit = a.FailedLimit
	r.FailedAbsorb = a.FailedAbsorb
	r.FailedReaction = a.FailedReaction
	r.InstructionStarved = a.InstructionStarved
	r.ActiveGenomes = make([]Genome, 0)
	for i, c := range w.Cells {
		for j, n := range c.Chemical {
			r.Chemicals[j] += int64(n)
		}
		if w.Config.Inflow*(w.Config.Width-i%w.Config.Width)/w.Config.Width > 0 {
			r.LitMatter += int64(c.Matter)
			if c.Matter == 0 {
				r.LitEmptyMatterCells++
			}
		} else {
			r.DarkMatter += int64(c.Matter)
		}
	}
	genomes := make(map[string]*Genome)
	lineages := make(map[string]int)
	var ages uint64
	for _, p := range w.Particles {
		ages += w.Tick - p.Created
		if len(p.Code) == 0 {
			continue
		}
		r.Executable++
		g := genomes[p.Genome]
		if g == nil {
			rec := w.Genomes[p.Genome]
			g = &Genome{Hash: p.Genome, FirstTick: rec.FirstTick, Births: rec.Births, Copies: rec.Copies, LastCopy: rec.LastCopy, Instructions: rec.Instructions, Moved: rec.Moved, Absorbed: rec.Absorbed, Converted: rec.Converted, Taken: rec.Taken, Transferred: rec.Transferred, Binds: rec.Binds}
			genomes[p.Genome] = g
		}
		g.Count++
		g.MaxGeneration = max(g.MaxGeneration, p.Generation)
		lineages[p.Origin]++
	}
	if r.Entities > 0 {
		r.MeanAge = float64(ages) / float64(r.Entities)
	}
	r.Genomes = len(genomes)
	r.Lineages = len(lineages)
	for _, g := range genomes {
		g.Frequency = float64(g.Count) / float64(r.Executable)
		r.ActiveGenomes = append(r.ActiveGenomes, *g)
	}
	sort.Slice(r.ActiveGenomes, func(i, j int) bool { return r.ActiveGenomes[i].Hash < r.ActiveGenomes[j].Hash })
	for hash, count := range lineages {
		r.ActiveLineages = append(r.ActiveLineages, Lineage{hash, count})
	}
	sort.Slice(r.ActiveLineages, func(i, j int) bool { return r.ActiveLineages[i].Hash < r.ActiveLineages[j].Hash })
	return r
}
