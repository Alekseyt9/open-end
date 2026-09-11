package experiment

import (
	"bytes"
	"fmt"
	"open-end/internal/kernel"
	"open-end/internal/rules"
	"open-end/internal/vm"
	"open-end/internal/world"
	"slices"
)

// RoleProtocol is external to physical snapshots. Replay starts from SourceHash
// with the same members, mode, donor, horizon and this protocol version.
type RoleProtocol struct {
	Version    int      `json:"version"`
	Mode       string   `json:"mode"`
	Donor      uint64   `json:"donor,omitempty"`
	Members    []uint64 `json:"members"`
	Ticks      int      `json:"ticks"`
	Every      int      `json:"every"`
	SwitchCost int      `json:"metabolic_switch_cost,omitempty"`
}
type CellRole struct {
	Founder           uint64               `json:"founder"`
	Genome            string               `json:"initial_genome"`
	AliveTicks        uint64               `json:"lineage_cell_ticks"`
	FounderAliveTicks uint64               `json:"founder_alive_ticks"`
	Alive             int                  `json:"lineage_alive"`
	Copies            uint64               `json:"successful_copies"`
	Allocations       uint64               `json:"successful_allocations"`
	Acquired          int64                `json:"acquired_energy"`
	Absorbed          int64                `json:"absorbed_energy"`
	Converted         int64                `json:"converted_energy"`
	Executed          map[vm.Opcode]uint64 `json:"paid_instructions_by_opcode"`
	Suppressed        map[vm.Opcode]uint64 `json:"suppressed_or_blinded_by_opcode"`
}
type RoleFlow struct {
	Source            uint64 `json:"source_founder"`
	Target            uint64 `json:"target_founder"`
	Kind              string `json:"kind"`
	Energy            int64  `json:"energy"`
	Events            uint64 `json:"events"`
	BondedEnergy      int64  `json:"bonded_energy"`
	ParentChildEnergy int64  `json:"observed_copy_parent_to_child_energy"`
}
type RoleFrame struct {
	Tick           uint64 `json:"tick"`
	Alive          []int  `json:"alive_by_founder"`
	FoundersAlive  int    `json:"founders_alive"`
	ConnectedPairs int    `json:"connected_founder_pairs"`
}
type RoleResult struct {
	Protocol              RoleProtocol         `json:"protocol"`
	SourceHash            string               `json:"source_sha256"`
	InitialHash           string               `json:"initial_sha256"`
	FinalHash             string               `json:"final_sha256"`
	Start                 uint64               `json:"start_tick"`
	FinalTick             uint64               `json:"final_tick"`
	RemovedBonds          int                  `json:"initial_removed_bonds"`
	Roles                 []CellRole           `json:"roles"`
	Flows                 []RoleFlow           `json:"flows"`
	Frames                []RoleFrame          `json:"frames"`
	TogetherTicks         uint64               `json:"all_original_members_connected_ticks"`
	ConnectedPairTicks    uint64               `json:"original_member_connected_pair_ticks"`
	StableDescendantTicks uint64               `json:"stable_founder_free_component_ticks"`
	Chemistry             *ChemicalTraceResult `json:"chemistry,omitempty"`
}
type roleSink struct {
	w           *world.World
	r           RoleResult
	tags        map[uint64]int // actual COPY ancestry, original member index + 1
	current     vm.Opcode
	flows       map[string]*RoleFlow
	ages        map[string]uint64
	copyParents map[uint64]uint64
	trace       *chemicalTrace
}

// BondComponents includes singletons, with sorted member lists. It is diagnostic
// only and consumes neither RNG nor mutable physical state.
func BondComponents(w *world.World) ([][]uint64, map[uint64]int) {
	adj := map[uint64][]uint64{}
	for _, b := range w.Relations {
		adj[b.A] = append(adj[b.A], b.B)
		adj[b.B] = append(adj[b.B], b.A)
	}
	ids := make([]uint64, 0, len(w.Particles))
	for id := range w.Particles {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	groups := [][]uint64{}
	lookup := map[uint64]int{}
	for _, id := range ids {
		if _, ok := lookup[id]; ok {
			continue
		}
		index := len(groups)
		group := []uint64{id}
		lookup[id] = index
		for i := 0; i < len(group); i++ {
			for _, q := range adj[group[i]] {
				if _, ok := lookup[q]; !ok {
					lookup[q] = index
					group = append(group, q)
				}
			}
		}
		slices.Sort(group)
		groups = append(groups, group)
	}
	return groups, lookup
}
func ValidateRoleMembers(w *world.World, members []uint64) error {
	if w == nil || len(members) < 2 || !slices.IsSorted(members) {
		return fmt.Errorf("members must be a sorted connected coded group")
	}
	groups, lookup := BondComponents(w)
	for i, id := range members {
		p := w.Particles[id]
		if p == nil || len(p.Code) == 0 || i > 0 && id == members[i-1] {
			return fmt.Errorf("invalid role member %d", id)
		}
	}
	if !slices.Equal(groups[lookup[members[0]]], members) {
		return fmt.Errorf("members are not an exact component")
	}
	return nil
}
func (s *roleSink) before(p *world.Particle, e *rules.Event) bool {
	s.current = e.Intent.Op
	index := s.tags[p.ID]
	if index == 0 {
		return false
	}
	r := &s.r.Roles[index-1]
	r.Executed[e.Intent.Op]++
	block, blind := false, false
	switch s.r.Protocol.Mode {
	case "no-sharing":
		block = e.Intent.Op == vm.TRANSFER && s.tags[p.Target] != 0
	case "no-peer-sharing":
		block = e.Intent.Op == vm.TRANSFER && s.tags[p.Target] != 0 && s.tags[p.Target] != index
	case "no-bonds":
		block = e.Intent.Op == vm.BIND && s.tags[p.Target] != 0
	case "anchored":
		block = e.Intent.Op == vm.MOVE
	case "no-bonds-anchored":
		block = e.Intent.Op == vm.MOVE || e.Intent.Op == vm.BIND && s.tags[p.Target] != 0
	case "no-reaction0", "no-reaction1":
		reaction := 0
		if s.r.Protocol.Mode == "no-reaction1" {
			reaction = 1
		}
		block = r.Founder == s.r.Protocol.Donor && e.Intent.Op == vm.CONVERT && e.Intent.A == reaction
	case "no-signals":
		block = e.Intent.Op == vm.EMIT || e.Intent.Op == vm.TOKEN
		blind = e.Intent.Op == vm.LISTEN || e.Intent.Op == vm.SENSE && (e.Intent.A == 7 || e.Intent.A >= 12 && e.Intent.A <= 15)
		if blind {
			e.Sensed = 0
			e.WordLength = 0
			e.ForeignWord = false
		}
	case "no-acquisition":
		block = r.Founder == s.r.Protocol.Donor && (e.Intent.Op == vm.ABSORB || e.Intent.Op == vm.CONVERT)
	}
	if block || blind {
		r.Suppressed[e.Intent.Op]++
	}
	return block
}
func (s *roleSink) Acquired(id uint64, genome string, energy int) {
	if i := s.tags[id]; i != 0 {
		r := &s.r.Roles[i-1]
		r.Acquired += int64(energy)
		if s.current == vm.ABSORB {
			r.Absorbed += int64(energy)
		}
		if s.current == vm.CONVERT {
			r.Converted += int64(energy)
		}
	}
}
func (s *roleSink) Interaction(e rules.Interaction) {
	a, b := s.tags[e.SourceID], s.tags[e.TargetID]
	if a != 0 {
		r := &s.r.Roles[a-1]
		if e.Kind == "copy" {
			s.tags[e.TargetID] = a
			if s.copyParents == nil {
				s.copyParents = map[uint64]uint64{}
			}
			s.copyParents[e.TargetID] = e.SourceID
			b = a
			r.Copies++
		}
		if e.Kind == "allocate" {
			r.Allocations++
		}
	}
	if (e.Kind != "transfer" && e.Kind != "take") || a == 0 && b == 0 {
		return
	}
	var source, target uint64
	if a != 0 {
		source = s.r.Roles[a-1].Founder
	}
	if b != 0 {
		target = s.r.Roles[b-1].Founder
	}
	kind := e.Kind
	if e.Target == "" {
		kind += "-uncoded-target"
	}
	key := fmt.Sprintf("%d:%d:%s", source, target, kind)
	f := s.flows[key]
	if f == nil {
		f = &RoleFlow{Source: source, Target: target, Kind: kind}
		s.flows[key] = f
	}
	f.Events++
	f.Energy += e.Energy
	if s.copyParents[e.TargetID] == e.SourceID {
		f.ParentChildEnergy += e.Energy
	}
	if e.DirectBond {
		f.BondedEnergy += e.Energy
	}
}
func (s *roleSink) Death(e rules.Death) { delete(s.tags, e.ID); delete(s.copyParents, e.ID) }
func (s *roleSink) frame(accumulate bool) RoleFrame {
	f := RoleFrame{Tick: s.w.Tick, Alive: make([]int, len(s.r.Roles))}
	groups, lookup := BondComponents(s.w)
	for id, i := range s.tags {
		if s.w.Particles[id] != nil {
			f.Alive[i-1]++
			if accumulate {
				s.r.Roles[i-1].AliveTicks++
			}
		}
	}
	for i := range s.r.Roles {
		r := &s.r.Roles[i]
		r.Alive = f.Alive[i]
		if s.w.Particles[r.Founder] != nil {
			f.FoundersAlive++
			if accumulate {
				r.FounderAliveTicks++
			}
			for j := 0; j < i; j++ {
				q := s.r.Roles[j].Founder
				if s.w.Particles[q] != nil && lookup[q] == lookup[r.Founder] {
					f.ConnectedPairs++
				}
			}
		}
	}
	if accumulate {
		s.r.ConnectedPairTicks += uint64(f.ConnectedPairs)
		if f.ConnectedPairs == len(s.r.Roles)*(len(s.r.Roles)-1)/2 {
			s.r.TogetherTicks++
		}
		eligible := map[string]bool{}
		for _, g := range groups {
			if len(g) < 2 {
				continue
			}
			valid := true
			for _, id := range g {
				if s.tags[id] == 0 || slices.Contains(s.r.Protocol.Members, id) {
					valid = false
					break
				}
			}
			if !valid {
				continue
			}
			key := fmt.Sprint(g)
			eligible[key] = true
			since, ok := s.ages[key]
			if !ok {
				since = s.w.Tick
				s.ages[key] = since
			}
			if s.w.Tick-since >= 100 {
				s.r.StableDescendantTicks++
			}
		}
		for k := range s.ages {
			if !eligible[k] {
				delete(s.ages, k)
			}
		}
	}
	return f
}
func (s *roleSink) TickCompleted(tick uint64) {
	f := s.frame(true)
	if (tick-s.r.Start)%uint64(s.r.Protocol.Every) == 0 || tick == s.r.Start+uint64(s.r.Protocol.Ticks) {
		s.r.Frames = append(s.r.Frames, f)
	}
}

func ContinueRoles(source *world.World, protocol RoleProtocol) (*world.World, RoleResult, error) {
	var empty RoleResult
	if (protocol.Version < 1 || protocol.Version > 3) || protocol.Ticks < 1 || protocol.Every < 1 || protocol.Every > protocol.Ticks {
		return nil, empty, fmt.Errorf("invalid role protocol")
	}
	if err := ValidateRoleMembers(source, protocol.Members); err != nil {
		return nil, empty, err
	}
	if source.Config.CollectiveAblation != "" || source.Config.BondMotion != "" {
		return nil, empty, fmt.Errorf("source already treated")
	}
	if protocol.Version >= 2 && (!source.Config.Ecology || source.RuleState != nil) {
		return nil, empty, fmt.Errorf("chemistry protocol requires baseline ecology without DSL state")
	}
	if protocol.Version == 3 {
		if source.Config.MetabolicSwitchCost != 0 || protocol.SwitchCost < 0 || protocol.SwitchCost > 64 || protocol.Mode != "intact" || protocol.Donor != 0 {
			return nil, empty, fmt.Errorf("switching role assay requires an untreated source and intact mode")
		}
	} else if protocol.SwitchCost != 0 {
		return nil, empty, fmt.Errorf("switch cost override requires protocol 3")
	}
	switch protocol.Mode {
	case "anchored", "no-bonds-anchored":
		if protocol.Version != 2 || protocol.Donor != 0 {
			return nil, empty, fmt.Errorf("anchoring requires protocol 2 and no donor")
		}
	case "no-reaction0", "no-reaction1":
		if protocol.Version != 2 || !slices.Contains(protocol.Members, protocol.Donor) {
			return nil, empty, fmt.Errorf("reaction ablation requires protocol 2 and a member donor")
		}
	case "intact", "no-sharing", "no-peer-sharing", "no-bonds", "no-signals":
		if protocol.Donor != 0 {
			return nil, empty, fmt.Errorf("unexpected donor")
		}
	case "no-acquisition":
		if !slices.Contains(protocol.Members, protocol.Donor) {
			return nil, empty, fmt.Errorf("donor must be a group member")
		}
	default:
		return nil, empty, fmt.Errorf("invalid role mode")
	}
	var buf bytes.Buffer
	if err := kernel.Save(&buf, source); err != nil {
		return nil, empty, err
	}
	w, err := kernel.Load(&buf)
	if err != nil {
		return nil, empty, err
	}
	if protocol.Version == 3 {
		w.Config.MetabolicSwitchCost = protocol.SwitchCost
	}
	protocol.Members = slices.Clone(protocol.Members)
	s := &roleSink{w: w, r: RoleResult{Protocol: protocol, SourceHash: kernel.Hash(source), Start: w.Tick}, tags: map[uint64]int{}, flows: map[string]*RoleFlow{}, ages: map[string]uint64{}}
	for i, id := range protocol.Members {
		s.tags[id] = i + 1
		s.r.Roles = append(s.r.Roles, CellRole{Founder: id, Genome: w.Particles[id].Genome, Executed: map[vm.Opcode]uint64{}, Suppressed: map[vm.Opcode]uint64{}})
	}
	if protocol.Version >= 2 {
		s.trace = newChemicalTrace(w, protocol.Members)
	}
	if protocol.Mode == "no-bonds" || protocol.Mode == "no-bonds-anchored" {
		for k, b := range w.Relations {
			if s.tags[b.A] != 0 && s.tags[b.B] != 0 {
				delete(w.Relations, k)
				s.r.RemovedBonds++
			}
		}
	}
	s.r.InitialHash = kernel.Hash(w)
	s.r.Frames = []RoleFrame{s.frame(false)}
	for i := 0; i < protocol.Ticks; i++ {
		kernel.StepWithIntervention(w, s, s.before)
	}
	if err := w.Validate(); err != nil {
		return nil, empty, err
	}
	s.r.FinalHash = kernel.Hash(w)
	s.r.FinalTick = w.Tick
	if s.trace != nil {
		if err := s.trace.finish(w); err != nil {
			return nil, empty, err
		}
		s.r.Chemistry = &s.trace.result
	}
	keys := make([]string, 0, len(s.flows))
	for k := range s.flows {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	s.r.Flows = []RoleFlow{}
	for _, k := range keys {
		s.r.Flows = append(s.r.Flows, *s.flows[k])
	}
	return w, s.r, nil
}
