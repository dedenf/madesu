package main

import (
	"reflect"
	"strings"
	"testing"
)

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

func TestLayoutChainTB(t *testing.T) {
	g := mustParse(t, "flowchart TD\nA --> B\nB --> C")
	lines, err := layoutFlowchart(g)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"[A]", " ▼", "[B]", " ▼", "[C]"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

func TestLayoutStandaloneNode(t *testing.T) {
	g := mustParse(t, "flowchart TD\nSolo[One]")
	lines, err := layoutFlowchart(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0] != "[One]" {
		t.Fatalf("got %q", lines)
	}
}

func TestLayoutShapes(t *testing.T) {
	g := mustParse(t, "flowchart TD\nA(Round)\nB{Why}\nC((Dot))")
	lines, err := layoutFlowchart(g)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"(Round)", "╱╲", "╱ Why ╲", "╲╱", "((Dot))"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in:\n%s", want, joined)
		}
	}
}

func TestLayoutLabelTruncated(t *testing.T) {
	long := strings.Repeat("x", 40)
	g := mustParse(t, "flowchart TD\nA["+long+"]")
	lines, err := layoutFlowchart(g)
	if err != nil {
		t.Fatal(err)
	}
	if got := len([]rune(lines[0])); got > 24+2 {
		t.Fatalf("box width %d exceeds 26: %q", got, lines[0])
	}
	if !strings.HasSuffix(lines[0], "…]") {
		t.Fatalf("no ellipsis: %q", lines[0])
	}
}

func TestLayoutStadiumLikeRounded(t *testing.T) {
	g := mustParse(t, "flowchart TD\nA([Stad])")
	lines, err := layoutFlowchart(g)
	if err != nil {
		t.Fatal(err)
	}
	if lines[0] != "(Stad)" {
		t.Fatalf("got %q, want \"(Stad)\"", lines[0])
	}
}
