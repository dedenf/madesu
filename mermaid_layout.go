package main

// Simplified layered layout for parsed flowcharts: break cycles, rank by
// longest path, order ranks with barycenter sweeps. Grid placement and edge
// routing live in the rest of this file (later tasks).

import (
	"sort"
	"strconv"
)

type box struct {
	n       *flowNode // nil for a virtual bend point
	virtual bool
	edgeIx  int
	rank    int
	order   int
	x, w, h int
}

type lchain struct {
	e      *flowEdge
	nodes  []*box // low rank → high rank
	hiIsTo bool   // the drawn arrow points at the high end
}

type layout struct {
	buckets [][]*box // per rank, in draw order
	chains  []*lchain
	rank    map[string]int
}

type workEdge struct {
	e   flowEdge
	rev bool // reversed for ranking because it closes a cycle
}

// breakCycles marks back edges via DFS so ranking sees an acyclic graph.
// Drawn edges keep the user's direction.
func breakCycles(g *flowGraph) []workEdge {
	adj := map[string][]int{}
	for i, e := range g.edges {
		adj[e.from] = append(adj[e.from], i)
	}
	const (
		white = 0
		gray  = 1
		black = 2
	)
	state := make(map[string]int, len(g.nodes))
	rev := make([]bool, len(g.edges))
	var dfs func(string)
	dfs = func(id string) {
		state[id] = gray
		for _, ei := range adj[id] {
			to := g.edges[ei].to
			switch state[to] {
			case white:
				dfs(to)
			case gray:
				rev[ei] = true
			}
		}
		state[id] = black
	}
	for _, n := range g.nodes {
		if state[n.id] == white {
			dfs(n.id)
		}
	}
	out := make([]workEdge, len(g.edges))
	for i, e := range g.edges {
		out[i] = workEdge{e: e, rev: rev[i]}
	}
	return out
}

// rankNodes assigns longest-path ranks over the acyclic working edge set.
func rankNodes(nodes []*flowNode, work []workEdge) map[string]int {
	indeg := map[string]int{}
	for _, w := range work {
		if !w.rev {
			indeg[w.e.to]++
		}
	}
	rank := map[string]int{}
	var q []string
	for _, n := range nodes {
		if indeg[n.id] == 0 {
			q = append(q, n.id)
		}
	}
	for len(q) > 0 {
		id := q[0]
		q = q[1:]
		for _, w := range work {
			if w.rev || w.e.from != id {
				continue
			}
			if r := rank[id] + 1; r > rank[w.e.to] {
				rank[w.e.to] = r
			}
			indeg[w.e.to]--
			if indeg[w.e.to] == 0 {
				q = append(q, w.e.to)
			}
		}
	}
	return rank
}

// planLayout builds boxes (one per node per rank, plus virtual bend points
// for edges spanning multiple ranks), chains, and an ordered bucket list.
func planLayout(g *flowGraph) (*layout, error) {
	for _, e := range g.edges {
		if e.from == e.to {
			return nil, errUnsupported
		}
	}
	if len(g.nodes) == 0 {
		return nil, errUnsupported
	}
	work := breakCycles(g)
	rank := rankNodes(g.nodes, work)

	byID := map[string]*flowNode{}
	for _, n := range g.nodes {
		byID[n.id] = n
	}
	key := func(id string, r int) string { return id + "\x00" + strconv.Itoa(r) }
	boxes := map[string]*box{}
	getBox := func(id string, r int) *box {
		k := key(id, r)
		b := boxes[k]
		if b == nil {
			b = &box{n: byID[id], rank: r}
			boxes[k] = b
		}
		return b
	}

	var chains []*lchain
	for i := range g.edges {
		e := &g.edges[i]
		lo, hi := e.from, e.to
		loR, hiR := rank[lo], rank[hi]
		hiIsTo := true
		if loR > hiR {
			lo, hi = hi, lo
			loR, hiR = hiR, loR
			hiIsTo = false
		}
		c := &lchain{e: e, hiIsTo: hiIsTo}
		for r := loR; r <= hiR; r++ {
			var b *box
			switch r {
			case loR:
				b = getBox(lo, r)
			case hiR:
				b = getBox(hi, r)
			default:
				b = &box{virtual: true, edgeIx: i, rank: r}
			}
			c.nodes = append(c.nodes, b)
		}
		chains = append(chains, c)
	}
	// Standalone definitions never mentioned by an edge still get a box.
	for _, n := range g.nodes {
		getBox(n.id, rank[n.id])
	}

	maxRank := 0
	for _, n := range g.nodes {
		if r := rank[n.id]; r > maxRank {
			maxRank = r
		}
	}
	buckets := make([][]*box, maxRank+1)
	for _, n := range g.nodes {
		b := boxes[key(n.id, rank[n.id])]
		buckets[b.rank] = append(buckets[b.rank], b)
	}
	for _, c := range chains {
		for _, b := range c.nodes {
			if b.virtual {
				buckets[b.rank] = append(buckets[b.rank], b)
			}
		}
	}
	for r := range buckets {
		for i, b := range buckets[r] {
			b.order = i
		}
	}
	orderBoxes(buckets, chains)
	return &layout{buckets: buckets, chains: chains, rank: rank}, nil
}

// orderBoxes runs four barycenter sweeps to reduce edge crossings. Stable
// sorting keeps ties in place, so identical input always yields one output.
func orderBoxes(buckets [][]*box, chains []*lchain) {
	nb := map[*box][]*box{}
	for _, c := range chains {
		for i, b := range c.nodes {
			if i > 0 {
				nb[b] = append(nb[b], c.nodes[i-1])
			}
			if i+1 < len(c.nodes) {
				nb[b] = append(nb[b], c.nodes[i+1])
			}
		}
	}
	bc := func(b *box, from int) float64 {
		sum, n := 0.0, 0.0
		for _, o := range nb[b] {
			if o.rank == from {
				sum += float64(o.order)
				n++
			}
		}
		if n == 0 {
			return float64(b.order)
		}
		return sum / n
	}
	for iter := 0; iter < 4; iter++ {
		if iter%2 == 0 {
			for r := 1; r < len(buckets); r++ {
				br := buckets[r]
				sort.SliceStable(br, func(i, j int) bool {
					return bc(br[i], r-1) < bc(br[j], r-1)
				})
				reindex(br)
			}
		} else {
			for r := len(buckets) - 2; r >= 0; r-- {
				br := buckets[r]
				sort.SliceStable(br, func(i, j int) bool {
					return bc(br[i], r+1) < bc(br[j], r+1)
				})
				reindex(br)
			}
		}
	}
}

func reindex(br []*box) {
	for i, b := range br {
		b.order = i
	}
}
