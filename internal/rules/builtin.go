// Package rules implements the replaceable baseline physics. The kernel owns
// scheduling and snapshots; these rules own instruction costs and local effects.
package rules

import (
	"open-end/internal/evolution"
	"open-end/internal/vm"
	"open-end/internal/world"
	"slices"
)

const Version = "ecology-2"

type Event struct {
	Actor  uint64
	Intent vm.Intent
	Sensed int
}

// Evaluate reads the tick-start view and emits an intent, without writing state.
func Evaluate(w *world.World, p *world.Particle) Event {
	intent := vm.Decode(p.Code, p.IP, p.Flag)
	sensed := p.Energy
	if intent.A == 1 {
		sensed = w.Cells[p.Position].Energy
	}
	if intent.A == 2 {
		sensed = w.Cells[p.Position].Matter
	}
	if intent.A >= 3 && intent.A <= 5 {
		sensed = w.Cells[p.Position].Chemical[intent.A-3]
	}
	return Event{Actor: p.ID, Intent: intent, Sensed: sensed}
}

// Resolve validates each intent against the state left by earlier resolutions.
func Resolve(w *world.World, e Event) {
	ResolveObserved(w, e, nil)
}
func ResolveObserved(w *world.World, e Event, sink Observer) {
	p := w.Particles[e.Actor]
	if p == nil {
		return
	}
	cost := 1
	switch e.Intent.Op {
	case vm.ALLOCATE:
		cost = 4
	case vm.COPY:
		cost = max(1, len(p.Code))
		if w.Config.CopyModel != "" && evolution.DecodePolicy(e.Intent.Instruction, w.Config.MutationPPM, w.Config.CopyModel == "evolving").Proofread {
			cost += 2
		}
	case vm.COPYMEM:
		cost = 8
	case vm.CONVERT:
		if w.Config.Ecology && w.RuleState != nil {
			if r := w.RuleState.Active.Find(e.Intent.A); r != nil {
				cost = r.EnergyCost
			}
		}
	}
	if p.Energy <= cost {
		w.Accounting.InstructionStarved++
		dissipate(w, p, p.Energy)
		return
	}
	dissipate(w, p, cost)
	p.IP = e.Intent.NextIP
	w.Accounting.Instructions++
	if r := w.Genomes[p.Genome]; r != nil {
		r.Instructions++
	}
	applyObserved(w, p, e, sink)
}

func Inflow(w *world.World) {
	for i := range w.Cells {
		rate := w.Config.Inflow * (w.Config.Width - i%w.Config.Width) / w.Config.Width
		n := min(rate, w.Config.CellCapacity-w.Cells[i].Energy)
		w.Cells[i].Energy += n
		w.Accounting.Injected += int64(n)
	}
	transport(w)
	if w.Config.Ecology {
		charge(w)
	}
}

func Decay(w *world.World, id uint64) {
	DecayObserved(w, id, nil)
}
func DecayObserved(w *world.World, id uint64, sink Observer) {
	p := w.Particles[id]
	dissipate(w, p, min(p.Energy, w.Config.Maintenance))
	if p.Energy == 0 {
		if sink != nil {
			sink.Death(Death{Tick: w.Tick + 1, Created: p.Created, Genome: p.Genome})
		}
		w.Unlink(id)
		c := &w.Cells[p.Position]
		c.Occupant = 0
		c.Matter++
		delete(w.Particles, id)
		w.Accounting.Deaths++
	}
}

func dissipate(w *world.World, p *world.Particle, n int) {
	p.Energy -= n
	w.Accounting.Dissipated += int64(n)
}
func target(w *world.World, p *world.Particle) *world.Particle {
	q := w.Particles[p.Target]
	if q == nil {
		return nil
	}
	for d := 0; d < 4; d++ {
		if w.Neighbor(p.Position, d) == q.Position {
			return q
		}
	}
	return nil
}

func destination(w *world.World, p *world.Particle, direction int, needMatter bool) int {
	start := direction
	if direction < 0 {
		start = w.RNG.Intn(4)
	}
	for n := 0; n < 4; n++ {
		pos := w.Neighbor(p.Position, start+n)
		if w.Cells[pos].Occupant == 0 && (!needMatter || w.Cells[pos].Matter > 0) {
			return pos
		}
		if direction >= 0 {
			break
		}
	}
	return -1
}

func apply(w *world.World, p *world.Particle, e Event) { applyObserved(w, p, e, nil) }

func applyObserved(w *world.World, p *world.Particle, e Event, sink Observer) {
	i := e.Intent.Instruction
	switch i.Op {
	case vm.NOP, vm.JUMP:
	case vm.SENSE:
		p.Memory[vm.Index(i.B, 8)] = e.Sensed
	case vm.WRITE:
		p.Memory[vm.Index(i.A, 8)] = i.B
	case vm.READ:
		p.Memory[vm.Index(i.B, 8)] = p.Memory[vm.Index(i.A, 8)]
	case vm.COMPARE:
		p.Flag = p.Memory[vm.Index(i.A, 8)] >= i.B
	case vm.MOVE:
		if w.Linked(p.ID) {
			return
		}
		pos := destination(w, p, i.A, false)
		if pos >= 0 {
			w.Cells[p.Position].Occupant = 0
			w.Cells[pos].Occupant = p.ID
			p.Position = pos
			if r := w.Genomes[p.Genome]; r != nil {
				r.Moved++
			}
		}
	case vm.ABSORB:
		c := &w.Cells[p.Position]
		n := max(0, min(i.A, c.Energy, w.Config.EnergyCapacity-p.Energy))
		c.Energy -= n
		p.Energy += n
		w.Accounting.Absorbed += int64(n)
		if n == 0 {
			w.Accounting.FailedAbsorb++
		}
		if r := w.Genomes[p.Genome]; r != nil {
			r.Absorbed += int64(n)
		}
	case vm.ALLOCATE:
		p.Target = 0
		if len(w.Particles) >= w.Config.MaxEntities {
			w.Accounting.FailedLimit++
			return
		}
		if p.Energy <= 12 {
			w.Accounting.FailedReserve++
			return
		}
		pos := destination(w, p, i.A, true)
		if pos < 0 {
			space := false
			for d := 0; d < 4; d++ {
				if i.A >= 0 && vm.Index(i.A, 4) != d {
					continue
				}
				space = space || w.Cells[w.Neighbor(p.Position, d)].Occupant == 0
			}
			if space {
				w.Accounting.FailedMatter++
			} else {
				w.Accounting.FailedSpace++
			}
			return
		}
		q := &world.Particle{ID: w.NextID, Parent: p.ID, Position: pos, Energy: 12, Created: w.Tick}
		q.Generation = p.Generation + 1
		w.NextID++
		p.Energy -= 12
		p.Target = q.ID
		w.Cells[pos].Matter--
		w.Cells[pos].Occupant = q.ID
		w.Particles[q.ID] = q
		w.Accounting.Allocations++
		emit(sink, w, "allocate", p, q, 12)
	case vm.COPYMEM:
		q := target(w, p)
		if q == nil || len(q.Code) != 0 {
			return
		}
		if w.Config.CopyModel == "" {
			q.Memory = p.Memory
			if w.RNG.Chance(w.Config.MutationPPM) {
				q.Memory[w.RNG.Intn(8)] = w.RNG.Intn(257) - 128
			}
		} else {
			policy := evolution.DecodePolicy(i, w.Config.MutationPPM, w.Config.CopyModel == "evolving")
			q.Memory = evolution.CopyMemory(p.Memory, p.InitialMemory, &w.RNG, policy)
			w.RecordCopy(p.Genome, policy, q.Memory != p.Memory, "")
		}
		q.InitialMemory = q.Memory
	case vm.COPY:
		q := target(w, p)
		if q == nil || len(q.Code) != 0 {
			return
		}
		ops := int(vm.BaselineOpcodeCount)
		if w.Config.Ecology {
			ops = int(vm.OpcodeCount)
		}
		if w.Config.CopyModel == "" {
			q.Code = evolution.MutateWithOpcodes(p.Code, &w.RNG, w.Config.MutationPPM, w.Config.MaxCode, ops)
		} else {
			policy := evolution.DecodePolicy(i, w.Config.MutationPPM, w.Config.CopyModel == "evolving")
			code, donor := p.Code, ""
			if policy.Recombine && w.RNG.Chance(policy.PPM) {
				start := w.RNG.Intn(4)
				for d := 0; d < 4; d++ {
					other := w.Particles[w.Cells[w.Neighbor(p.Position, start+d)].Occupant]
					if other != nil && other.ID != p.ID && other.ID != q.ID && len(other.Code) > 0 {
						code = evolution.Recombine(code, other.Code, &w.RNG)
						donor = other.Genome
						break
					}
				}
			}
			q.Code = evolution.MutatePolicy(code, &w.RNG, policy, w.Config.MaxCode, ops)
			w.RecordCopy(p.Genome, policy, !slices.Equal(q.Code, p.Code), donor)
		}
		w.RegisterGenome(q, p.Genome)
		if r := w.Genomes[p.Genome]; r != nil {
			r.Copies++
			r.LastCopy = w.Tick
		}
		q.Origin = evolution.Hash(q.Code, q.InitialMemory)
		o := w.Origins[q.Origin]
		if o == nil {
			o = &world.Origin{Hash: q.Origin, Parent: p.Origin, FirstTick: w.Tick}
			w.Origins[q.Origin] = o
		}
		o.Copies++
		w.Accounting.Copies++
		emit(sink, w, "copy", p, q, 0)
	case vm.TRANSFER:
		q := target(w, p)
		if q == nil {
			return
		}
		n := max(0, min(i.A, p.Energy-1, w.Config.EnergyCapacity-q.Energy))
		p.Energy -= n
		q.Energy += n
		w.Accounting.Transferred += int64(n)
		if n > 0 {
			emit(sink, w, "transfer", p, q, n)
		}
		if r := w.Genomes[p.Genome]; r != nil {
			r.Transferred += int64(n)
		}
	case vm.CONVERT:
		if w.Config.Ecology {
			convert(w, p, i.A, i.B)
		}
	case vm.TARGET:
		p.Target = 0
		start := i.A
		if start < 0 {
			start = w.RNG.Intn(4)
		}
		for d := 0; d < 4; d++ {
			id := w.Cells[w.Neighbor(p.Position, start+d)].Occupant
			if id != 0 {
				p.Target = id
				break
			}
			if i.A >= 0 {
				break
			}
		}
	case vm.TAKE:
		q := target(w, p)
		if q == nil {
			return
		}
		n := max(0, min(i.A, q.Energy, w.Config.EnergyCapacity-p.Energy))
		q.Energy -= n
		p.Energy += n
		w.Accounting.Taken += int64(n)
		if n > 0 {
			emit(sink, w, "take", q, p, n)
		} // direction is resource flow, victim -> actor
		if r := w.Genomes[p.Genome]; r != nil {
			r.Taken += int64(n)
		}
	case vm.BIND:
		q := target(w, p)
		if q == nil {
			return
		}
		key := world.RelationKey(p.ID, q.ID)
		if _, exists := w.Relations[key]; !exists {
			w.Relations[key] = world.Relation{A: min(p.ID, q.ID), B: max(p.ID, q.ID)}
			emit(sink, w, "bind", p, q, 0)
			if r := w.Genomes[p.Genome]; r != nil {
				r.Binds++
			}
		}
	case vm.UNBIND:
		if i.A < 0 {
			unlinkObserved(sink, w, p)
		} else {
			if _, ok := w.Relations[world.RelationKey(p.ID, p.Target)]; ok {
				emit(sink, w, "unbind", p, w.Particles[p.Target], 0)
			}
			delete(w.Relations, world.RelationKey(p.ID, p.Target))
		}
	}
}
