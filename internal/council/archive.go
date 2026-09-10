package council

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"open-end/internal/kernel"
	"open-end/internal/observer"
	"open-end/internal/world"
)

// ArchiveOptions describes research selection heuristics, never world fitness.
type ArchiveOptions struct {
	MinTicks       uint64  `json:"min_ticks"`
	MinSurvival    float64 `json:"min_survival_fraction"`
	MinPersistence float64 `json:"min_productive_block_fraction"`
	Bins           int     `json:"behavior_bins_per_octave"`
	Neighbors      int     `json:"novelty_neighbors"`
}

func DefaultArchiveOptions() ArchiveOptions { return ArchiveOptions{10000, .75, .75, 1, 3} }

var archiveFeatures = []string{"allocate", "copy", "transfer", "take", "bind", "unbind", "absorbed_energy"}

type ArchivePoint struct {
	Node        string    `json:"node"`
	RequestID   string    `json:"request_sha256"`
	Context     string    `json:"comparison_context"`
	Horizon     string    `json:"horizon_group"`
	Cell        string    `json:"behavior_cell"`
	Descriptor  []float64 `json:"log_behavior_descriptor"`
	CellBins    []int     `json:"cell_bins"`
	Novelty     *float64  `json:"novelty_distance"`
	Nearest     []string  `json:"novelty_reference_nodes"`
	DominatedBy []string  `json:"dominated_by,omitempty"`
	Diversity   float64   `json:"mean_effective_diversity"`
	Structure   float64   `json:"mean_bonded_scale_proxy"`
	Persistence float64   `json:"productive_block_fraction"`
	Survival    float64   `json:"surviving_world_fraction"`
	Eligible    bool      `json:"eligible"`
	Tip         bool      `json:"tip"`
	Pareto      bool      `json:"pareto_tip"`
	Protected   bool      `json:"cell_representative"`
	Recommended bool      `json:"recommended"`
	Reasons     []string  `json:"reasons"`
}
type ArchiveDecision struct {
	Version     int            `json:"version"`
	ID          string         `json:"archive_sha256"`
	TreeHash    string         `json:"tree_nodes_sha256"`
	Options     ArchiveOptions `json:"options"`
	Features    []string       `json:"behavior_features"`
	Unavailable []string       `json:"unavailable_metrics"`
	Points      []ArchivePoint `json:"points"`
	Recommended []string       `json:"recommended"`
}

type archiveContext struct {
	Case     string
	Seed     uint64
	Config   world.Config
	Duration uint64
	Samples  int
	Blocks   []uint64
	Detector observer.DynamicsConfig
}

func validArchiveOptions(o ArchiveOptions) bool {
	fraction := func(x float64) bool { return !math.IsNaN(x) && x > 0 && x <= 1 }
	return o.MinTicks > 0 && fraction(o.MinSurvival) && fraction(o.MinPersistence) && o.Bins >= 1 && o.Bins <= 16 && o.Neighbors >= 1 && o.Neighbors <= 32
}

// archivePoint uses common physical actions only. Reaction IDs and DSL hashes
// are intentionally excluded: their meanings can change between rule modules.
func archivePoint(dir string, n TreeNode, o ArchiveOptions) (ArchivePoint, error) {
	p := ArchivePoint{Node: n.ID, RequestID: n.RequestID, Descriptor: []float64{}, CellBins: []int{}, Nearest: []string{}, Reasons: []string{}}
	checked, err := checkRound(nodeRound(dir, n.ID), "", true)
	if err != nil {
		return p, err
	}
	worlds := append([]WorldBrief(nil), checked.Request.Worlds...)
	sort.Slice(worlds, func(i, j int) bool {
		if worlds[i].Case != worlds[j].Case {
			return worlds[i].Case < worlds[j].Case
		}
		return worlds[i].Seed < worlds[j].Seed
	})
	contexts := []archiveContext{}
	horizons := []uint64{}
	means := make([]float64, len(archiveFeatures))
	ready := true
	seen := map[string]bool{}
	for _, wb := range worlds {
		key := fmt.Sprintf("%s/%d", wb.Case, wb.Seed)
		if seen[key] {
			return p, fmt.Errorf("duplicate case/seed in cohort")
		}
		seen[key] = true
		w, err := loadWorld(filepath.Join(nodeRound(dir, n.ID), wb.ID+".snapshot.json"))
		if err != nil {
			return p, err
		}
		if kernel.Hash(w) != wb.SnapshotHash {
			return p, fmt.Errorf("source snapshot changed during archive validation")
		}
		var e Evidence
		data, err := os.ReadFile(filepath.Join(nodeRound(dir, n.ID), wb.ID+".evidence.json"))
		if err != nil {
			return p, err
		}
		if digest(data) != wb.EvidenceHash {
			return p, fmt.Errorf("source evidence changed during archive validation")
		}
		if err = decode(data, &e); err != nil {
			return p, err
		}
		s, d := e.Summary, e.Dynamics
		ctx := archiveContext{wb.Case, wb.Seed, w.Config, s.ToTick - s.FromTick, s.Samples, []uint64{}, d.Config}
		for _, b := range d.Behavior {
			ctx.Blocks = append(ctx.Blocks, b.ToTick-b.FromTick)
		}
		contexts = append(contexts, ctx)
		horizons = append(horizons, wb.Tick)
		alive := s.DiversityEnd.EffectiveGenomes > 0
		if alive {
			p.Survival++
		}
		p.Diversity += s.DiversityEnd.EffectiveGenomes
		// Particle-weighted log component size: singletons contribute zero.
		scale := 0.0
		for _, size := range s.StructuresEnd.Sizes {
			if size.Size > 0 {
				scale += float64(size.Size) * float64(size.Count) * math.Log2(float64(size.Size))
			}
		}
		p.Structure += scale / float64(max(int64(1), s.Population.End))
		if ctx.Duration < o.MinTicks || len(d.Behavior) != 4 || len(s.RuleEvents) != 0 || s.RulesStart != s.RulesEnd {
			ready = false
			p.Reasons = append(p.Reasons, wb.ID+": insufficient stable four-block history")
			continue
		}
		indices := []int{}
		for _, name := range archiveFeatures {
			idx := -1
			for i, f := range d.Features {
				if f == name {
					if idx != -1 {
						return p, fmt.Errorf("duplicate behavior feature")
					}
					idx = i
				}
			}
			if idx < 0 {
				return p, fmt.Errorf("missing common behavior feature %s", name)
			}
			indices = append(indices, idx)
		}
		for _, block := range d.Behavior {
			for _, index := range indices {
				if index >= len(block.Rates) || block.Rates[index] < 0 || math.IsNaN(block.Rates[index]) || math.IsInf(block.Rates[index], 0) {
					return p, fmt.Errorf("invalid behavior rate")
				}
			}
			if alive && block.Rates[indices[1]] > 0 {
				p.Persistence += .25
			}
		}
		for j, index := range indices {
			value := math.Log2(1 + (d.Behavior[2].Rates[index]+d.Behavior[3].Rates[index])/2)
			p.Descriptor = append(p.Descriptor, value)
			means[j] += value
		}
	}
	count := float64(len(worlds))
	p.Diversity /= count
	p.Structure /= count
	p.Persistence /= count
	p.Survival /= count
	p.Context = jsonHash(contexts)
	p.Horizon = jsonHash(struct {
		Context string
		Ticks   []uint64
	}{p.Context, horizons})
	p.Eligible = ready && p.Survival >= o.MinSurvival && p.Persistence >= o.MinPersistence
	if p.Survival < o.MinSurvival {
		p.Reasons = append(p.Reasons, "survival filter")
	}
	if p.Persistence < o.MinPersistence {
		p.Reasons = append(p.Reasons, "productive-block persistence filter")
	}
	if !ready {
		p.Descriptor = nil
		return p, nil
	}
	for _, value := range means {
		p.CellBins = append(p.CellBins, int(math.Floor(value/count*float64(o.Bins))))
	}
	p.Cell = jsonHash(struct {
		Context string
		Bins    []int
	}{p.Context, p.CellBins})
	return p, nil
}

func descriptorDistance(a, b []float64) float64 {
	total := 0.0
	for i := range a {
		total += math.Abs(a[i] - b[i])
	}
	return total / float64(len(a))
}

// Pareto dominance requires no worse values on every available objective and
// at least one strict improvement. Equal vectors never eliminate each other.
func archiveDominates(a, b ArchivePoint) bool {
	av := []float64{a.Diversity, a.Structure, a.Persistence}
	bv := []float64{b.Diversity, b.Structure, b.Persistence}
	if a.Novelty != nil && b.Novelty != nil {
		av = append(av, *a.Novelty)
		bv = append(bv, *b.Novelty)
	}
	better := false
	for i := range av {
		if av[i] < bv[i] {
			return false
		}
		if av[i] > bv[i] {
			better = true
		}
	}
	return better
}

// rankArchive is deterministic and independent of filesystem enumeration.
// Novelty is recomputed relative to each decision's complete reference set;
// older immutable decisions preserve what was known at that time.
func rankArchive(points []ArchivePoint, o ArchiveOptions) []ArchivePoint {
	sort.Slice(points, func(i, j int) bool { return points[i].Node < points[j].Node })
	type neighbor struct {
		node     string
		distance float64
	}
	for i := range points {
		p := &points[i]
		p.Novelty = nil
		p.Nearest = []string{}
		p.DominatedBy = []string{}
		p.Pareto = false
		p.Protected = false
		p.Recommended = false
		if !p.Eligible {
			continue
		}
		peers := []neighbor{}
		seenDescriptors := map[string]bool{}
		exact := ""
		for j, q := range points {
			if j == i || !q.Eligible || q.Context != p.Context {
				continue
			}
			distance := descriptorDistance(p.Descriptor, q.Descriptor)
			if distance == 0 {
				exact = q.Node
				break
			}
			key := jsonHash(q.Descriptor)
			if seenDescriptors[key] {
				continue
			}
			seenDescriptors[key] = true
			peers = append(peers, neighbor{q.Node, distance})
		}
		if exact != "" {
			zero := 0.0
			p.Novelty = &zero
			p.Nearest = []string{exact}
			continue
		}
		sort.Slice(peers, func(i, j int) bool {
			if peers[i].distance != peers[j].distance {
				return peers[i].distance < peers[j].distance
			}
			return peers[i].node < peers[j].node
		})
		if len(peers) > 0 {
			sum := 0.0
			for _, q := range peers[:min(o.Neighbors, len(peers))] {
				sum += q.distance
				p.Nearest = append(p.Nearest, q.node)
			}
			distance := sum / float64(len(p.Nearest))
			p.Novelty = &distance
		}
	}
	activeContexts := map[string]bool{}
	for i := range points {
		p := &points[i]
		if !p.Eligible || !p.Tip {
			continue
		}
		activeContexts[p.Context] = true
		p.Pareto = true
		for j, q := range points {
			if i != j && q.Eligible && q.Tip && q.Horizon == p.Horizon && archiveDominates(q, *p) {
				p.Pareto = false
				p.DominatedBy = append(p.DominatedBy, q.Node)
			}
		}
		if len(p.DominatedBy) > 0 {
			p.Reasons = append(p.Reasons, "dominated at matched horizon by "+strings.Join(p.DominatedBy, ","))
		}
	}
	// Preserve the oldest representative of each observed viable behavior cell.
	// Cell retention deliberately does not require global Pareto membership.
	representatives := map[string]int{}
	for i, p := range points {
		if p.Eligible {
			if _, ok := representatives[p.Cell]; !ok {
				representatives[p.Cell] = i
			}
		}
	}
	covered := map[string]bool{}
	for i := range points {
		if points[i].Pareto {
			points[i].Recommended = true
			covered[points[i].Cell] = true
			points[i].Reasons = append(points[i].Reasons, "non-dominated tip at matched horizon")
		}
	}
	for _, i := range representatives {
		points[i].Protected = true
	}
	// Prefer a viable tip in an uncovered cell, so ordinary retention keeps
	// moving forward. Revisit its historical representative only if necessary.
	cells := make([]string, 0, len(representatives))
	for cell := range representatives {
		cells = append(cells, cell)
	}
	sort.Strings(cells)
	for _, cell := range cells {
		i := representatives[cell]
		if covered[cell] || !activeContexts[points[i].Context] {
			continue
		}
		for j, q := range points {
			if q.Cell == cell && q.Eligible && q.Tip {
				i = j
				break
			}
		}
		points[i].Recommended = true
		points[i].Reasons = append(points[i].Reasons, "retained representative of an otherwise uncovered behavior cell")
		covered[cell] = true
	}
	return points
}

func BuildArchive(dir string, v TreeView, o ArchiveOptions) (ArchiveDecision, error) {
	a := ArchiveDecision{Version: 1, TreeHash: jsonHash(v.Nodes), Options: o, Features: append([]string{}, archiveFeatures...), Unavailable: []string{"causal_structure", "hierarchy_depth", "new_information_processing", "adaptive_value"}, Points: []ArchivePoint{}, Recommended: []string{}}
	if !validArchiveOptions(o) {
		return a, fmt.Errorf("invalid archive options")
	}
	parents := map[string]bool{}
	for _, n := range v.Nodes {
		parents[n.Parent] = true
	}
	for _, n := range v.Nodes {
		p, err := archivePoint(dir, n, o)
		if err != nil {
			return a, fmt.Errorf("%s: %w", n.ID, err)
		}
		p.Tip = !parents[n.ID]
		a.Points = append(a.Points, p)
	}
	a.Points = rankArchive(a.Points, o)
	for _, p := range a.Points {
		if p.Recommended {
			a.Recommended = append(a.Recommended, p.Node)
		}
	}
	a.ID = jsonHash(a)
	return a, nil
}

func readArchiveDecision(path string) (ArchiveDecision, error) {
	var a ArchiveDecision
	if err := readJSON(path, &a); err != nil {
		return a, err
	}
	id := a.ID
	a.ID = ""
	if a.Version != 1 || !validArchiveOptions(a.Options) || jsonHash(a) != id {
		return a, fmt.Errorf("invalid archive identity")
	}
	a.ID = id
	return a, nil
}

// LatestArchive returns nil when a tree predates Stage 9. Incomplete revision
// directories are ignored; corrupt published decisions are never ignored.
func LatestArchive(dir string) (*ArchiveDecision, error) {
	path := filepath.Join(dir, "archive")
	entries, err := os.ReadDir(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if !entries[i].IsDir() {
			continue
		}
		file := filepath.Join(path, entries[i].Name(), "decision.json")
		a, err := readArchiveDecision(file)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return &a, nil
	}
	return nil, nil
}

// ArchiveTree records a new immutable decision only when inputs/options change.
// Applying affects only the saved continuation selection, never world physics.
func ArchiveTree(dir string, o ArchiveOptions, apply bool) (ArchiveDecision, error) {
	unlock, err := treeLock(dir)
	if err != nil {
		return ArchiveDecision{}, err
	}
	defer unlock()
	v, err := ReadTree(dir)
	if err != nil {
		return ArchiveDecision{}, err
	}
	a, err := BuildArchive(dir, v, o)
	if err != nil {
		return a, err
	}
	old, err := LatestArchive(dir)
	if err != nil {
		return a, err
	}
	if old == nil || old.ID != a.ID {
		path := filepath.Join(dir, "archive")
		if err = os.MkdirAll(path, 0755); err != nil {
			return a, err
		}
		id, err := reserve(path, "a")
		if err != nil {
			return a, err
		}
		if err = replaceJSON(filepath.Join(path, id, "decision.json"), a); err != nil {
			return a, err
		}
	}
	if apply {
		if len(a.Recommended) == 0 {
			return a, fmt.Errorf("no eligible recommendations; previous selection retained")
		}
		if err = replaceJSON(filepath.Join(dir, "selection.json"), a.Recommended); err != nil {
			return a, err
		}
	}
	return a, nil
}
func ArchiveText(a ArchiveDecision) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Archive %s\nNode             Eligible  Tip    Pareto  Cell rep  Novelty  Diversity  Structure  Persistence\n", a.ID)
	for _, p := range a.Points {
		novelty := "n/a"
		if p.Novelty != nil {
			novelty = fmt.Sprintf("%.3f", *p.Novelty)
		}
		fmt.Fprintf(&b, "%-16s %-9t %-6t %-7t %-9t %-8s %-10.3f %-10.3f %.3f\n", p.Node, p.Eligible, p.Tip, p.Pareto, p.Protected, novelty, p.Diversity, p.Structure, p.Persistence)
	}
	fmt.Fprintf(&b, "Recommended: %s\nUnavailable: %s\n", strings.Join(a.Recommended, ","), strings.Join(a.Unavailable, ", "))
	return b.String()
}
