package discovery

import (
	"open-end/internal/world"
	"slices"
)

// Strong components require a positive TRANSFER path in both directions.
// TAKE, unprogrammed particles, and dead endpoints cannot create such a boundary.
func reciprocal(w *world.World, edges map[pair]edge) [][]uint64 {
	adj := map[uint64][]uint64{}
	live := func(id uint64) bool { p := w.Particles[id]; return p != nil && len(p.Code) > 0 }
	for p, e := range edges {
		if e.transfer > 0 && live(p.a) && live(p.b) {
			adj[p.a] = append(adj[p.a], p.b)
			if _, ok := adj[p.b]; !ok {
				adj[p.b] = nil
			}
		}
	}
	ids := []uint64{}
	for id := range adj {
		ids = append(ids, id)
		slices.Sort(adj[id])
	}
	slices.Sort(ids)
	index, low := map[uint64]int{}, map[uint64]int{}
	on := map[uint64]bool{}
	stack := []uint64{}
	next := 0
	result := [][]uint64{}
	var visit func(uint64)
	visit = func(v uint64) {
		next++
		index[v] = next
		low[v] = next
		stack = append(stack, v)
		on[v] = true
		for _, u := range adj[v] {
			if index[u] == 0 {
				visit(u)
				low[v] = min(low[v], low[u])
			} else if on[u] {
				low[v] = min(low[v], index[u])
			}
		}
		if low[v] == index[v] {
			members := []uint64{}
			for {
				u := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				on[u] = false
				members = append(members, u)
				if u == v {
					break
				}
			}
			if len(members) > 1 {
				slices.Sort(members)
				result = append(result, members)
			}
		}
	}
	for _, id := range ids {
		if index[id] == 0 {
			visit(id)
		}
	}
	slices.SortFunc(result, func(a, b []uint64) int { return slices.Compare(a, b) })
	return result
}
func reciprocalWithin(w *world.World, edges map[pair]edge, ids []uint64) bool {
	set := map[uint64]bool{}
	for _, id := range ids {
		set[id] = true
	}
	induced := map[pair]edge{}
	for p, e := range edges {
		if set[p.a] && set[p.b] {
			induced[p] = e
		}
	}
	groups := reciprocal(w, induced)
	return len(groups) == 1 && slices.Equal(groups[0], ids)
}
func contains(a, b []uint64) bool { // a is a strict superset of b; IDs are sorted.
	if len(a) <= len(b) {
		return false
	}
	i := 0
	for _, v := range a {
		if v == b[i] {
			i++
			if i == len(b) {
				return true
			}
		}
	}
	return false
}
func intersects(a, b []uint64) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			return true
		}
		if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return false
}

// Minimal inclusion DAG, never forced parents for overlapping boundaries.
func hierarchy(nodes []Node) {
	for i := range nodes {
		for j := i - 1; j >= 0; j-- {
			if !contains(nodes[i].Members, nodes[j].Members) {
				continue
			}
			covered := false
			for _, id := range nodes[i].Children {
				for k := j + 1; k < i; k++ {
					if nodes[k].ID == id && contains(nodes[k].Members, nodes[j].Members) {
						covered = true
						break
					}
				}
				if covered {
					break
				}
			}
			if !covered {
				nodes[i].Children = append(nodes[i].Children, nodes[j].ID)
				nodes[i].Level = max(nodes[i].Level, nodes[j].Level+1)
			}
		}
		slices.Sort(nodes[i].Children)
		for j := 0; j < i; j++ {
			if intersects(nodes[i].Members, nodes[j].Members) && !contains(nodes[i].Members, nodes[j].Members) {
				nodes[i].Overlap++
				nodes[j].Overlap++
			}
		}
	}
}
