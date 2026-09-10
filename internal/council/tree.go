package council

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TreeNode struct {
	ID         string      `json:"id"`
	Parent     string      `json:"parent,omitempty"`
	Variant    string      `json:"variant"`
	Generation int         `json:"generation"`
	RequestID  string      `json:"request_sha256"`
	Run        string      `json:"run,omitempty"`
	Created    string      `json:"created_utc"`
	Worlds     []TreeWorld `json:"worlds"`
}
type TreeWorld struct {
	CollectiveAblation     string  `json:"collective_ablation,omitempty"`
	CollectiveAge          uint64  `json:"collective_age,omitempty"`
	Environment            string  `json:"environment,omitempty"`
	CopyModel              string  `json:"copy_model,omitempty"`
	CopyPolicies           int     `json:"copy_policies,omitempty"`
	PersistentCopyPolicies int     `json:"persistent_copy_policies,omitempty"`
	ID                     string  `json:"id"`
	Case                   string  `json:"case"`
	Seed                   uint64  `json:"seed"`
	From                   uint64  `json:"from_tick"`
	Tick                   uint64  `json:"tick"`
	Population             int64   `json:"population"`
	Diversity              float64 `json:"diversity"`
	Largest                int     `json:"largest_structure"`
	Copies                 uint64  `json:"copies"`
	Status                 string  `json:"status"`
	SnapshotHash           string  `json:"snapshot_sha256"`
	RulesHash              string  `json:"rules_sha256"`
}
type TreeRun struct {
	ID       string       `json:"id"`
	Parents  []string     `json:"parents"`
	Children []string     `json:"children"`
	Status   string       `json:"status"`
	Error    string       `json:"error,omitempty"`
	Options  TrialOptions `json:"options"`
	Seconds  float64      `json:"seconds"`
}
type TreeView struct {
	Version  int        `json:"version"`
	Nodes    []TreeNode `json:"nodes"`
	Selected []string   `json:"selected"`
	Runs     []TreeRun  `json:"runs"`
}

// All mutations hold an exclusive lock. A crashed process leaves the lock
// visible for operator recovery; it must never be silently stolen.
func treeLock(dir string) (func(), error) {
	p := filepath.Join(dir, ".write.lock")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("tree is locked or unavailable: %w", err)
	}
	_, err = fmt.Fprintf(f, "pid=%d started=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339))
	closeErr := f.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return func() { _ = os.Remove(p) }, nil
}
func replaceJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".publish-*")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, err = f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
func nodeRound(dir, id string) string { return filepath.Join(dir, "nodes", id, "round") }
func reserve(dir, prefix string) (string, error) {
	for i := 1; i <= 999999; i++ {
		id := fmt.Sprintf("%s%06d", prefix, i)
		err := os.Mkdir(filepath.Join(dir, id), 0755)
		if err == nil {
			return id, nil
		}
		if !os.IsExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("directory ID space exhausted")
}
func commitNode(dir, id, parent, variant, run string, generation int, r Request) (TreeNode, error) {
	n := TreeNode{ID: id, Parent: parent, Variant: variant, Run: run, Generation: generation, RequestID: r.ID, Created: time.Now().UTC().Format(time.RFC3339), Worlds: []TreeWorld{}}
	for _, w := range r.Worlds {
		var e Evidence
		if err := readJSON(filepath.Join(nodeRound(dir, id), w.ID+".evidence.json"), &e); err != nil {
			return n, err
		}
		n.Worlds = append(n.Worlds, treeWorld(w, e))
	}
	return n, replaceJSON(filepath.Join(dir, "nodes", id, "node.json"), n)
}

func treeWorld(w WorldBrief, e Evidence) TreeWorld {
	r := TreeWorld{ID: w.ID, Case: w.Case, Seed: w.Seed, From: e.Summary.FromTick, Tick: w.Tick, Population: e.Summary.Population.End, Diversity: e.Summary.DiversityEnd.EffectiveGenomes, Largest: e.Summary.StructuresEnd.Largest, Copies: e.Summary.Copies, Status: e.Dynamics.Status, SnapshotHash: w.SnapshotHash, RulesHash: w.Rules.Hash}
	r.CollectiveAblation = w.CollectiveAblation
	if e.Summary.Collectives != nil {
		r.CollectiveAge = e.Summary.Collectives.End.MinAge
	}
	if v := e.Summary.Variation; v != nil {
		r.CopyModel = v.Model
		r.CopyPolicies = v.CodePolicies
		r.PersistentCopyPolicies = v.PersistentCodePolicies
	}
	if v := e.Summary.Environment; v != nil {
		r.Environment = v.Model
	}
	return r
}

// InitTree freezes the source evidence and snapshots. Moving or deleting the
// original experiment cannot affect future tree continuations.
func InitTree(input, dir, variant string, window uint64) error {
	if window == 0 {
		return fmt.Errorf("window must be positive")
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}
	unlock, err := treeLock(dir)
	if err != nil {
		return err
	}
	defer unlock()
	for _, part := range []string{"nodes", "runs"} {
		if err := os.Mkdir(filepath.Join(dir, part), 0755); err != nil {
			return err
		}
	}
	id, err := reserve(filepath.Join(dir, "nodes"), "b")
	if err != nil {
		return err
	}
	r, err := PrepareVariant(input, nodeRound(dir, id), window, variant)
	if err != nil {
		return err
	}
	if _, err = commitNode(dir, id, "", "source", "", 0, r); err != nil {
		return err
	}
	if err = writeJSON(filepath.Join(dir, "selection.json"), []string{id}); err != nil {
		return err
	}
	// Publishing tree.json last is the initialization completion gate.
	return writeJSON(filepath.Join(dir, "tree.json"), map[string]int{"version": 1})
}

func ReadTree(dir string) (TreeView, error) {
	v := TreeView{Nodes: []TreeNode{}, Selected: []string{}, Runs: []TreeRun{}}
	var header struct {
		Version int `json:"version"`
	}
	if err := readJSON(filepath.Join(dir, "tree.json"), &header); err != nil {
		return v, err
	}
	if header.Version != 1 {
		return v, fmt.Errorf("unsupported tree version")
	}
	v.Version = header.Version
	entries, err := os.ReadDir(filepath.Join(dir, "nodes"))
	if err != nil {
		return v, err
	}
	known := map[string]TreeNode{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, "nodes", entry.Name(), "node.json")
		var n TreeNode
		if err := readJSON(path, &n); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return v, err
		}
		if n.ID != entry.Name() || !identifier.MatchString(n.ID) || len(n.Worlds) == 0 {
			return v, fmt.Errorf("invalid node")
		}
		var r Request
		if err := readJSON(filepath.Join(nodeRound(dir, n.ID), "request.json"), &r); err != nil {
			return v, err
		}
		id := r.ID
		r.ID = ""
		if id != n.RequestID || jsonHash(r) != id || len(n.Worlds) != len(r.Worlds) {
			return v, fmt.Errorf("node/request identity mismatch")
		}
		for i, w := range r.Worlds {
			if !identifier.MatchString(w.ID) {
				return v, fmt.Errorf("invalid world ID")
			}
			b, err := os.ReadFile(filepath.Join(nodeRound(dir, n.ID), w.ID+".evidence.json"))
			if err != nil {
				return v, err
			}
			if digest(b) != w.EvidenceHash {
				return v, fmt.Errorf("node evidence changed")
			}
			var e Evidence
			if err := decode(b, &e); err != nil {
				return v, err
			}
			if n.Worlds[i] != treeWorld(w, e) {
				return v, fmt.Errorf("node world metrics or identity mismatch")
			}
		}
		known[n.ID] = n
		v.Nodes = append(v.Nodes, n)
	}
	roots := 0
	for _, n := range v.Nodes {
		if n.Parent == "" {
			roots++
			if n.Generation != 0 {
				return v, fmt.Errorf("invalid root generation")
			}
			continue
		}
		p, ok := known[n.Parent]
		if !ok || n.Generation != p.Generation+1 {
			return v, fmt.Errorf("invalid ancestry for %s", n.ID)
		}
		if !identifier.MatchString(n.Run) {
			return v, fmt.Errorf("invalid run ID")
		}
		if err := validateTreeEdge(dir, p, n); err != nil {
			return v, err
		}
	}
	if roots != 1 {
		return v, fmt.Errorf("tree requires exactly one root")
	}
	if err := readJSON(filepath.Join(dir, "selection.json"), &v.Selected); err != nil {
		return v, err
	}
	if err := validateSelection(v, v.Selected); err != nil {
		return v, err
	}
	entries, err = os.ReadDir(filepath.Join(dir, "runs"))
	if err != nil {
		return v, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		var r TreeRun
		if err := readJSON(filepath.Join(dir, "runs", e.Name(), "run.json"), &r); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return v, err
		}
		if r.ID != e.Name() {
			return v, fmt.Errorf("invalid run identity")
		}
		v.Runs = append(v.Runs, r)
	}
	return v, nil
}

func validateTreeEdge(dir string, parent, child TreeNode) error {
	path := filepath.Join(dir, "runs", child.Run, parent.ID)
	var manifest trialManifest
	if err := readJSON(filepath.Join(path, "manifest.json"), &manifest); err != nil {
		return err
	}
	if manifest.Status != "complete" || manifest.RequestID != parent.RequestID || manifest.Completed != manifest.Total {
		return fmt.Errorf("incomplete or unrelated parent trial")
	}
	var rows []TrialRow
	if err := readJSON(filepath.Join(path, "results.json"), &rows); err != nil {
		return err
	}
	if len(rows) != manifest.Total {
		return fmt.Errorf("trial result count mismatch")
	}
	key := func(c string, seed uint64) string { return fmt.Sprintf("%s/%d", c, seed) }
	parents := map[string]TreeWorld{}
	for _, w := range parent.Worlds {
		parents[key(w.Case, w.Seed)] = w
	}
	matched := map[string]bool{}
	for _, row := range rows {
		if row.Variant != child.Variant {
			continue
		}
		k := key(row.Case, row.Seed)
		p, ok := parents[k]
		if !ok || matched[k] || row.SourceHash != p.SnapshotHash || row.Error != "" {
			return fmt.Errorf("invalid branch origin")
		}
		found := false
		for _, w := range child.Worlds {
			if key(w.Case, w.Seed) == k && w.SnapshotHash == row.FinalHash && w.RulesHash == row.RulesHash {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("child does not match trial results")
		}
		matched[k] = true
	}
	if len(matched) != len(parent.Worlds) || len(matched) != len(child.Worlds) {
		return fmt.Errorf("branch cohort is incomplete")
	}
	return nil
}
func validateSelection(v TreeView, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("select at least one branch")
	}
	known := map[string]bool{}
	for _, n := range v.Nodes {
		known[n.ID] = true
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if !known[id] || seen[id] {
			return fmt.Errorf("unknown or duplicate branch %q", id)
		}
		seen[id] = true
	}
	return nil
}
func SelectTree(dir string, ids []string) error {
	unlock, err := treeLock(dir)
	if err != nil {
		return err
	}
	defer unlock()
	v, err := ReadTree(dir)
	if err != nil {
		return err
	}
	if err = validateSelection(v, ids); err != nil {
		return err
	}
	sort.Strings(ids)
	return replaceJSON(filepath.Join(dir, "selection.json"), ids)
}

// GrowTree advances all selected cohorts. Proposal mode requires a checked
// response.json for each parent; continuation mode performs no AI invocation.
// Cohorts are scheduled sequentially, each sharing the same bounded worker pool
// size, so total CPU concurrency never multiplies by the number of parents.
func GrowTree(dir string, ids []string, proposals bool, opts TrialOptions, progress io.Writer) (children []string, err error) {
	if opts.Ticks < 1 || opts.Every < 1 || opts.Workers < 1 || opts.Workers > 256 || opts.Window == 0 || opts.Window > uint64(opts.Ticks) {
		return nil, fmt.Errorf("positive ticks/every/window required, window <= ticks, workers 1..256")
	}
	unlock, err := treeLock(dir)
	if err != nil {
		return nil, err
	}
	defer unlock()
	v, err := ReadTree(dir)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		ids = append([]string(nil), v.Selected...)
	}
	if err = validateSelection(v, ids); err != nil {
		return nil, err
	}
	sort.Strings(ids)
	nodes := map[string]TreeNode{}
	for _, n := range v.Nodes {
		nodes[n.ID] = n
	}
	checked := map[string]Checked{}
	// Validate every parent before starting any compute or creating a run.
	for _, id := range ids {
		round := nodeRound(dir, id)
		response := ""
		if proposals {
			response = filepath.Join(round, "response.json")
		}
		c, e := checkRound(round, response, !proposals)
		if e != nil {
			return nil, fmt.Errorf("%s: %w", id, e)
		}
		if proposals && len(c.Modules) == 0 {
			return nil, fmt.Errorf("%s: no proposals; use continuation mode", id)
		}
		for _, w := range c.Request.Worlds {
			if uint64(opts.Ticks) > ^uint64(0)-w.Tick {
				return nil, fmt.Errorf("tick overflow")
			}
		}
		checked[id] = c
	}
	runID, err := reserve(filepath.Join(dir, "runs"), "r")
	if err != nil {
		return nil, err
	}
	runPath := filepath.Join(dir, "runs", runID)
	record := TreeRun{ID: runID, Parents: ids, Children: []string{}, Status: "running", Options: opts}
	recordPath := filepath.Join(runPath, "run.json")
	if err = writeJSON(recordPath, record); err != nil {
		return nil, err
	}
	start := time.Now()
	defer func() {
		record.Seconds = time.Since(start).Seconds()
		record.Children = append([]string{}, children...)
		record.Status = "complete"
		if err != nil {
			record.Status = "failed"
			record.Error = err.Error()
		}
		if e := replaceJSON(recordPath, record); e != nil && err == nil {
			err = e
		}
	}()
	for _, id := range ids {
		c := checked[id]
		trialPath := filepath.Join(runPath, id)
		if _, err = executeTrial(c, nodeRound(dir, id), trialPath, opts, progress); err != nil {
			return children, err
		}
		variants := []string{"control"}
		for _, p := range c.Response.Proposals {
			variants = append(variants, p.ID)
		}
		for _, variant := range variants {
			child, e := reserve(filepath.Join(dir, "nodes"), "b")
			if e != nil {
				return children, e
			}
			request, e := PrepareVariant(trialPath, nodeRound(dir, child), opts.Window, variant)
			if e != nil {
				return children, e
			}
			if _, e = commitNode(dir, child, id, variant, runID, nodes[id].Generation+1, request); e != nil {
				return children, e
			}
			children = append(children, child)
			record.Children = append([]string{}, children...)
			if err = replaceJSON(recordPath, record); err != nil {
				return children, err
			}
		}
	}
	// Keep every child eligible; selection never deletes older branches.
	err = replaceJSON(filepath.Join(dir, "selection.json"), children)
	return children, err
}
func TreeText(v TreeView) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Branches: %d | selected: %s\n", len(v.Nodes), strings.Join(v.Selected, ","))
	for _, n := range v.Nodes {
		var population int64
		var diversity float64
		for _, w := range n.Worlds {
			population += w.Population
			diversity += w.Diversity
		}
		fmt.Fprintf(&b, "%s%s <- %s | %s | %d worlds | population %d | mean diversity %.3f\n", strings.Repeat("  ", n.Generation), n.ID, n.Parent, n.Variant, len(n.Worlds), population, diversity/float64(len(n.Worlds)))
	}
	for _, r := range v.Runs {
		fmt.Fprintf(&b, "%s: %s, %d children, %.2fs %s\n", r.ID, r.Status, len(r.Children), r.Seconds, r.Error)
	}
	return b.String()
}
