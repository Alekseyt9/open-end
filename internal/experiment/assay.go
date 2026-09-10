// Package experiment contains experimental interventions, outside world physics.
package experiment

import (
	"fmt"
	"open-end/internal/evolution"
	"open-end/internal/kernel"
	"open-end/internal/vm"
	"open-end/internal/world"
)

type Template struct {
	Hash string           `json:"hash"`
	Code []vm.Instruction `json:"code"`
}
type Pair struct {
	SourceHash string   `json:"source_state_sha256"`
	SourceTick uint64   `json:"source_tick"`
	SourceSeed uint64   `json:"source_seed"`
	A          Template `json:"a"`
	B          Template `json:"b"`
}

func Extract(w *world.World, a, b string) (Pair, error) {
	p := Pair{SourceHash: kernel.Hash(w), SourceTick: w.Tick, SourceSeed: w.Config.Seed}
	for hash, dest := range map[string]*Template{a: &p.A, b: &p.B} {
		r := w.Genomes[hash]
		if r == nil {
			return p, fmt.Errorf("unknown genome %s", hash)
		}
		active := false
		for _, particle := range w.Particles {
			active = active || particle.Genome == hash
		}
		if !active {
			return p, fmt.Errorf("genome is not alive in source")
		}
		*dest = Template{Hash: hash, Code: append([]vm.Instruction(nil), r.Code...)}
	}
	return p, p.Validate()
}

func (p Pair) Validate() error {
	if p.A.Hash == p.B.Hash {
		return fmt.Errorf("select two distinct genomes")
	}
	for _, g := range []Template{p.A, p.B} {
		if len(g.Code) == 0 || len(g.Code) > 64 || g.Hash != evolution.Hash(g.Code, [8]int{}) {
			return fmt.Errorf("invalid template hash or size")
		}
		for _, i := range g.Code {
			if i.Op >= vm.OpcodeCount {
				return fmt.Errorf("invalid template opcode")
			}
		}
	}
	return nil
}

var AssayCases = []string{"a-only", "b-only", "a-half", "b-half", "mixed", "mixed-no-chemical-diffusion"}

// Inoculate standardizes energy, memory and environment. At a given seed all
// treatments use the same shuffled positions. Even positions belong to A in
// mixed and a-half; odd positions belong to B in mixed and b-half.
func Inoculate(pair Pair, seed uint64, condition string, founders int) (*world.World, error) {
	if err := pair.Validate(); err != nil {
		return nil, err
	}
	if founders < 2 || founders > 1024 || founders%2 != 0 {
		return nil, fmt.Errorf("founders must be even and in [2,1024]")
	}
	found := false
	for _, c := range AssayCases {
		found = found || condition == c
	}
	if !found {
		return nil, fmt.Errorf("unknown assay condition")
	}
	c := world.DefaultConfig()
	c.Width = 32
	c.Height = 32
	c.MaxEntities = 1024
	c.Ecology = true
	c.MutationPPM = 0
	c.MatterDiffusion = 4
	c.ChemicalDiffusion = 4
	c.Seed = seed
	if condition == "mixed-no-chemical-diffusion" {
		c.ChemicalDiffusion = 0
	}
	w, err := world.New(c)
	if err != nil {
		return nil, err
	}
	// Remove only the constructor's seed; this defines a fresh initial state.
	for _, p := range w.Particles {
		w.Cells[p.Position].Occupant = 0
		w.Cells[p.Position].Matter++
	}
	w.Particles = make(map[uint64]*world.Particle)
	w.Genomes = make(map[string]*world.GenomeRecord)
	w.Origins = make(map[string]*world.Origin)
	w.NextID = 1
	positions := make([]int, len(w.Cells))
	for i := range positions {
		positions[i] = i
	}
	placement := evolution.RNG{State: seed ^ 0xbb67ae8584caa73b}
	for i := len(positions) - 1; i > 0; i-- {
		j := placement.Intn(i + 1)
		positions[i], positions[j] = positions[j], positions[i]
	}
	for i, pos := range positions[:founders] {
		if condition == "a-half" && i%2 != 0 || condition == "b-half" && i%2 == 0 {
			continue
		}
		g := pair.A
		if condition == "b-only" || condition == "b-half" || (condition == "mixed" || condition == "mixed-no-chemical-diffusion") && i%2 != 0 {
			g = pair.B
		}
		p := &world.Particle{ID: w.NextID, Position: pos, Energy: 128, Code: append([]vm.Instruction(nil), g.Code...)}
		w.NextID++
		p.Origin = evolution.Hash(p.Code, p.InitialMemory)
		if w.Origins[p.Origin] == nil {
			w.Origins[p.Origin] = &world.Origin{Hash: p.Origin}
		}
		w.Origins[p.Origin].Copies++
		w.RegisterGenome(p, "")
		w.Particles[p.ID] = p
		w.Cells[pos].Occupant = p.ID
		w.Cells[pos].Matter--
	}
	e, m := w.Totals()
	w.Accounting = world.Accounting{InitialEnergy: e, InitialMatter: m, InitialChemical: int64(len(w.Cells) * 32)}
	return w, w.Validate()
}
