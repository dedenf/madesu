package main

import "testing"

func TestRankChainAndDiamond(t *testing.T) {
	g := mustParse(t, "flowchart TD\nA --> B\nB --> C")
	rank := rankNodes(g.nodes, breakCycles(g))
	if rank["A"] != 0 || rank["B"] != 1 || rank["C"] != 2 {
		t.Fatalf("chain ranks %v", rank)
	}
	g = mustParse(t, "flowchart TD\nA --> B\nA --> C\nB --> D\nC --> D")
	rank = rankNodes(g.nodes, breakCycles(g))
	if rank["A"] != 0 || rank["B"] != 1 || rank["C"] != 1 || rank["D"] != 2 {
		t.Fatalf("diamond ranks %v", rank)
	}
}

func TestRankCycle(t *testing.T) {
	g := mustParse(t, "flowchart TD\nA --> B\nB --> C\nC --> A")
	work := breakCycles(g)
	rev := 0
	for _, w := range work {
		if w.rev {
			rev++
		}
	}
	if rev != 1 {
		t.Fatalf("reversed %d edges, want 1", rev)
	}
	rank := rankNodes(g.nodes, work)
	if rank["A"] != 0 || rank["B"] != 1 || rank["C"] != 2 {
		t.Fatalf("cycle ranks %v", rank)
	}
}

func TestPlanLayoutDeterministic(t *testing.T) {
	src := "flowchart TD\nA --> B\nA --> C\nB --> D\nC --> D"
	l1, err := planLayout(mustParse(t, src))
	if err != nil {
		t.Fatal(err)
	}
	l2, err := planLayout(mustParse(t, src))
	if err != nil {
		t.Fatal(err)
	}
	for r := range l1.buckets {
		for i, b := range l1.buckets[r] {
			o := l2.buckets[r][i]
			if (b.n == nil) != (o.n == nil) || (b.n != nil && o.n.id != b.n.id) {
				t.Fatalf("bucket %d pos %d differs: %v vs %v", r, i, b.n, o.n)
			}
		}
	}
}

func TestPlanLayoutBarycenter(t *testing.T) {
	// A→D and B→C cross under first-seen order; one sweep uncrosses them.
	g := mustParse(t, "flowchart TD\nA --> D\nB --> C")
	l, err := planLayout(g)
	if err != nil {
		t.Fatal(err)
	}
	if l.buckets[1][0].n.id != "D" || l.buckets[1][1].n.id != "C" {
		t.Fatalf("rank1 order %s %s, want D C",
			l.buckets[1][0].n.id, l.buckets[1][1].n.id)
	}
}

func TestPlanLayoutSelfLoopRejected(t *testing.T) {
	g := mustParse(t, "flowchart TD\nA --> A")
	if _, err := planLayout(g); err == nil {
		t.Fatal("self loop accepted")
	}
}
