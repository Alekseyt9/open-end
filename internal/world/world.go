// Package world contains physical state and reproducibility metadata, without
// biological classes or a fitness function. Energy uses integer units.
package world

import (
	"fmt"
	"open-end/internal/dsl"
	"open-end/internal/evolution"
	"open-end/internal/vm"
)

type Config struct {
	Width             int    `json:"width"`
	Height            int    `json:"height"`
	MaxEntities       int    `json:"max_entities"`
	MaxCode           int    `json:"max_code"`
	CellCapacity      int    `json:"cell_capacity"`
	EnergyCapacity    int    `json:"energy_capacity"`
	Inflow            int    `json:"inflow"`
	Maintenance       int    `json:"maintenance"`
	MutationPPM       int    `json:"mutation_ppm"`
	Seed              uint64 `json:"seed"`
	MatterDiffusion   int    `json:"matter_diffusion"`
	Ecology           bool   `json:"ecology"`
	ChemicalDiffusion int    `json:"chemical_diffusion"`
	CopyModel         string `json:"copy_model,omitempty"`
	Environment       string `json:"environment,omitempty"`
}

func DefaultConfig() Config {
	return Config{Width: 256, Height: 256, MaxEntities: 65536, MaxCode: 64,
		CellCapacity: 128, EnergyCapacity: 256, Inflow: 4, Maintenance: 1,
		MutationPPM: 10000, Seed: 1}
}

func (c Config) Validate() error {
	if c.Environment != "" && c.Environment != "coupled" && c.Environment != "inert" {
		return fmt.Errorf("environment must be empty, coupled, or inert")
	}
	if c.Environment != "" && !c.Ecology {
		return fmt.Errorf("environment requires ecology")
	}
	if c.CopyModel != "" && c.CopyModel != "fixed" && c.CopyModel != "evolving" {
		return fmt.Errorf("copy_model must be empty, fixed, or evolving")
	}
	if c.Width < 2 || c.Height < 2 || c.Width > 1024 || c.Height > 1024 {
		return fmt.Errorf("grid dimensions must be in [2,1024]")
	}
	if c.MaxEntities < 1 || c.MaxEntities > c.Width*c.Height {
		return fmt.Errorf("max_entities must be in [1, width*height]")
	}
	if c.MaxCode < len(vm.Seed()) || c.MaxCode > 1024 {
		return fmt.Errorf("max_code must be in [%d,1024]", len(vm.Seed()))
	}
	if c.Ecology && c.MaxCode < len(vm.EcologySeed()) {
		return fmt.Errorf("max_code is too small for ecology seed")
	}
	if c.CellCapacity < 1 || c.CellCapacity > 1_000_000 || c.EnergyCapacity < 128 || c.EnergyCapacity > 1_000_000 {
		return fmt.Errorf("invalid energy capacities")
	}
	if c.Inflow < 0 || c.Inflow > c.CellCapacity || c.Maintenance < 0 || c.Maintenance > c.EnergyCapacity {
		return fmt.Errorf("invalid inflow or maintenance")
	}
	if c.MutationPPM < 0 || c.MutationPPM > 1_000_000 {
		return fmt.Errorf("mutation_ppm must be in [0,1000000]")
	}
	if c.MatterDiffusion < 0 || c.MatterDiffusion > 1000000 || c.ChemicalDiffusion < 0 || c.ChemicalDiffusion > 1000000 {
		return fmt.Errorf("diffusion intervals must be in [0,1000000]")
	}
	return nil
}

type Cell struct {
	Terrain  int    `json:"terrain,omitempty"`
	Signal   int    `json:"signal,omitempty"`
	Energy   int    `json:"energy"`
	Matter   int    `json:"matter"`
	Occupant uint64 `json:"occupant"`
	Chemical [3]int `json:"chemical"`
}

type Particle struct {
	ID       uint64           `json:"id"`
	Parent   uint64           `json:"parent"`
	Position int              `json:"position"`
	Energy   int              `json:"energy"`
	Code     []vm.Instruction `json:"code"`
	Memory   [8]int           `json:"memory"`
	// InitialMemory distinguishes inherited state from execution scratch space.
	InitialMemory [8]int `json:"initial_memory"`
	IP            int    `json:"ip"`
	Flag          bool   `json:"flag"`
	Target        uint64 `json:"target"`
	Created       uint64 `json:"created"`
	Origin        string `json:"origin"`
	Genome        string `json:"genome"`
	Generation    uint64 `json:"generation"`
}

// Origin records a copy provenance edge. Multiple origins can share code;
// identity includes inherited memory. It has no effect on physics.
type Origin struct {
	Hash      string `json:"hash"`
	Parent    string `json:"parent"`
	FirstTick uint64 `json:"first_tick"`
	Copies    uint64 `json:"copies"`
}

type Accounting struct {
	InitialEnergy      int64    `json:"initial_energy"`
	Injected           int64    `json:"injected"`
	Dissipated         int64    `json:"dissipated"`
	InitialMatter      int64    `json:"initial_matter"`
	Allocations        uint64   `json:"allocations"`
	Copies             uint64   `json:"copies"`
	Deaths             uint64   `json:"deaths"`
	Instructions       uint64   `json:"instructions"`
	Absorbed           int64    `json:"absorbed"`
	Transferred        int64    `json:"transferred"`
	InitialChemical    int64    `json:"initial_chemical"`
	Converted          [2]int64 `json:"converted"`
	Charged            int64    `json:"charged"`
	Taken              int64    `json:"taken"`
	FailedSpace        uint64   `json:"failed_space"`
	FailedMatter       uint64   `json:"failed_matter"`
	FailedReserve      uint64   `json:"failed_reserve"`
	FailedLimit        uint64   `json:"failed_limit"`
	FailedAbsorb       uint64   `json:"failed_absorb"`
	FailedReaction     uint64   `json:"failed_reaction"`
	InstructionStarved uint64   `json:"instruction_starved"`
}

// GenomeRecord is a cumulative observation ledger, never consulted by physics.
type GenomeRecord struct {
	Hash         string           `json:"hash"`
	Code         []vm.Instruction `json:"code"`
	Parent       string           `json:"parent"`
	FirstTick    uint64           `json:"first_tick"`
	Births       uint64           `json:"births"`
	Copies       uint64           `json:"copies"`
	LastCopy     uint64           `json:"last_copy_tick"`
	Instructions uint64           `json:"instructions"`
	Moved        uint64           `json:"moved"`
	Absorbed     int64            `json:"absorbed"`
	Converted    [2]int64         `json:"converted"`
	Transferred  int64            `json:"transferred"`
	Taken        int64            `json:"taken"`
	Binds        uint64           `json:"binds"`
}

type Relation struct {
	A uint64 `json:"a"`
	B uint64 `json:"b"`
}

type World struct {
	Config       Config                           `json:"config"`
	Tick         uint64                           `json:"tick"`
	RNG          evolution.RNG                    `json:"rng"`
	NextID       uint64                           `json:"next_id"`
	Cells        []Cell                           `json:"cells"`
	Particles    map[uint64]*Particle             `json:"particles"`
	Origins      map[string]*Origin               `json:"origins"`
	Accounting   Accounting                       `json:"accounting"`
	TransportRNG evolution.RNG                    `json:"transport_rng"`
	Genomes      map[string]*GenomeRecord         `json:"genomes"`
	Relations    map[string]Relation              `json:"relations"`
	RuleState    *dsl.State                       `json:"rule_state,omitempty"`
	Variation    map[string]*evolution.CopyRecord `json:"variation,omitempty"`
	Environment  *EnvironmentState                `json:"environment,omitempty"`
}

func New(c Config) (*World, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	w := &World{Config: c, RNG: evolution.RNG{State: c.Seed}, NextID: 2,
		Cells: make([]Cell, c.Width*c.Height), Particles: make(map[uint64]*Particle), Origins: make(map[string]*Origin)}
	w.TransportRNG = evolution.RNG{State: c.Seed ^ 0x6a09e667f3bcc909}
	w.Genomes = make(map[string]*GenomeRecord)
	if c.Environment != "" {
		w.Environment = &EnvironmentState{}
	}
	if c.CopyModel != "" {
		w.Variation = make(map[string]*evolution.CopyRecord)
	}
	w.Relations = make(map[string]Relation)
	for i := range w.Cells {
		w.Cells[i] = Cell{Energy: c.CellCapacity / 2, Matter: 1}
		if c.Ecology {
			w.Cells[i].Chemical = [3]int{16, 0, 16}
		}
	}
	pos := (c.Height/2)*c.Width + c.Width/2
	p := &Particle{ID: 1, Position: pos, Energy: 128, Code: vm.Seed()}
	if c.Ecology {
		p.Code = vm.EcologySeed()
	}
	p.Origin = evolution.Hash(p.Code, p.InitialMemory)
	w.Particles[1] = p
	w.Origins[p.Origin] = &Origin{Hash: p.Origin, Copies: 1}
	w.RegisterGenome(p, "")
	w.Cells[pos].Matter--
	w.Cells[pos].Occupant = 1
	w.Accounting.InitialEnergy = int64(len(w.Cells)*(c.CellCapacity/2) + p.Energy)
	w.Accounting.InitialMatter = int64(len(w.Cells))
	if c.Ecology {
		w.Accounting.InitialChemical = int64(len(w.Cells) * 32)
		w.Accounting.InitialEnergy += int64(len(w.Cells) * 16 * 8)
	}
	return w, nil
}

// Neighbor is a toroidal von Neumann neighborhood, direction N/E/S/W.
func (w *World) Neighbor(pos, direction int) int {
	x, y := pos%w.Config.Width, pos/w.Config.Width
	switch vm.Index(direction, 4) {
	case 0:
		y--
	case 1:
		x++
	case 2:
		y++
	case 3:
		x--
	}
	return vm.Index(y, w.Config.Height)*w.Config.Width + vm.Index(x, w.Config.Width)
}

func (w *World) Totals() (energy, matter int64) {
	for _, c := range w.Cells {
		energy += int64(c.Energy)
		energy += int64(c.Signal)
		energy += int64(8*c.Chemical[0] + 4*c.Chemical[1])
		matter += int64(c.Matter)
		matter += int64(c.Terrain)
	}
	for _, p := range w.Particles {
		energy += int64(p.Energy)
		matter++
	}
	return
}

func (w *World) RegisterGenome(p *Particle, parent string) {
	p.Genome = evolution.Hash(p.Code, [8]int{})
	r := w.Genomes[p.Genome]
	if r == nil {
		r = &GenomeRecord{Hash: p.Genome, Code: append([]vm.Instruction(nil), p.Code...), Parent: parent, FirstTick: w.Tick}
		w.Genomes[p.Genome] = r
	}
	r.Births++
}

func RelationKey(a, b uint64) string {
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("%d:%d", a, b)
}

func (w *World) Linked(id uint64) bool {
	if len(w.Relations) == 0 {
		return false
	}
	p := w.Particles[id]
	if p == nil {
		return false
	}
	for d := 0; d < 4; d++ {
		q := w.Cells[w.Neighbor(p.Position, d)].Occupant
		if _, ok := w.Relations[RelationKey(id, q)]; ok {
			return true
		}
	}
	return false
}

func (w *World) Unlink(id uint64) {
	if len(w.Relations) == 0 {
		return
	}
	p := w.Particles[id]
	if p == nil {
		return
	}
	for d := 0; d < 4; d++ {
		q := w.Cells[w.Neighbor(p.Position, d)].Occupant
		delete(w.Relations, RelationKey(id, q))
	}
}

func (w *World) Validate() error {
	if err := w.ValidateEnvironment(); err != nil {
		return err
	}
	if err := w.ValidateVariation(); err != nil {
		return err
	}
	if w.RuleState != nil && !w.Config.Ecology {
		return fmt.Errorf("DSL state requires ecology")
	}
	if err := w.RuleState.Validate(w.Tick); err != nil {
		return err
	}
	if err := w.Config.Validate(); err != nil {
		return err
	}
	if len(w.Cells) != w.Config.Width*w.Config.Height || len(w.Particles) > w.Config.MaxEntities || w.Particles == nil || w.Origins == nil || w.Genomes == nil || w.Relations == nil {
		return fmt.Errorf("invalid world storage")
	}
	var chemical int64
	for i, c := range w.Cells {
		for _, n := range c.Chemical {
			if n < 0 || n > len(w.Cells)*32 {
				return fmt.Errorf("invalid chemical at %d", i)
			}
			chemical += int64(n)
		}
		if c.Energy < 0 || c.Energy > w.Config.CellCapacity || c.Matter < 0 || c.Matter > len(w.Cells) {
			return fmt.Errorf("invalid cell %d", i)
		}
		if c.Occupant != 0 {
			p := w.Particles[c.Occupant]
			if p == nil || p.Position != i {
				return fmt.Errorf("invalid occupancy at %d", i)
			}
		}
	}
	for id, p := range w.Particles {
		if p == nil || id == 0 || p.ID != id || id >= w.NextID || p.Position < 0 || p.Position >= len(w.Cells) {
			return fmt.Errorf("invalid particle %d", id)
		}
		if w.Cells[p.Position].Occupant != id || p.Energy <= 0 || p.Energy > w.Config.EnergyCapacity || len(p.Code) > w.Config.MaxCode || p.Created > w.Tick {
			return fmt.Errorf("invalid particle state %d", id)
		}
		if len(p.Code) > 0 {
			if p.Genome != evolution.Hash(p.Code, [8]int{}) || w.Genomes[p.Genome] == nil {
				return fmt.Errorf("invalid genome ledger for %d", id)
			}
			if p.IP < 0 || p.IP >= len(p.Code) || p.Origin != evolution.Hash(p.Code, p.InitialMemory) || w.Origins[p.Origin] == nil {
				return fmt.Errorf("invalid program provenance %d", id)
			}
		}
		for _, i := range p.Code {
			if i.Op >= vm.OpcodeCount || w.Config.Environment == "" && i.Op >= vm.EcologyOpcodeCount {
				return fmt.Errorf("invalid opcode for %d", id)
			}
		}
	}
	if chemical != w.Accounting.InitialChemical {
		return fmt.Errorf("chemical matter accounting mismatch")
	}
	for hash, r := range w.Genomes {
		if r == nil || r.Hash != hash || r.Hash != evolution.Hash(r.Code, [8]int{}) || len(r.Code) == 0 || len(r.Code) > w.Config.MaxCode || r.FirstTick > w.Tick || r.LastCopy > w.Tick {
			return fmt.Errorf("invalid genome record")
		}
		for _, i := range r.Code {
			if i.Op >= vm.OpcodeCount || w.Config.Environment == "" && i.Op >= vm.EcologyOpcodeCount {
				return fmt.Errorf("invalid archived opcode")
			}
		}
		if r.Parent != "" && w.Genomes[r.Parent] == nil {
			return fmt.Errorf("missing genome parent")
		}
	}
	for hash, o := range w.Origins {
		if o == nil || o.Hash != hash || o.FirstTick > w.Tick || o.Parent != "" && w.Origins[o.Parent] == nil {
			return fmt.Errorf("invalid origin record")
		}
	}
	for key, b := range w.Relations {
		p, q := w.Particles[b.A], w.Particles[b.B]
		if p == nil || q == nil || b.A >= b.B || key != RelationKey(b.A, b.B) {
			return fmt.Errorf("invalid relation")
		}
		near := false
		for d := 0; d < 4; d++ {
			near = near || w.Neighbor(p.Position, d) == q.Position
		}
		if !near {
			return fmt.Errorf("nonlocal relation")
		}
	}
	e, m := w.Totals()
	if e != w.Accounting.InitialEnergy+w.Accounting.Injected-w.Accounting.Dissipated || m != w.Accounting.InitialMatter {
		return fmt.Errorf("resource accounting mismatch")
	}
	return nil
}
