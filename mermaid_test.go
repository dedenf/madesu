package main

import (
	"errors"
	"testing"
)

func mustParse(t *testing.T, src string) *flowGraph {
	t.Helper()
	g, err := parseFlowchart([]byte(src))
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return g
}

func TestParseHeader(t *testing.T) {
	cases := []struct {
		src  string
		want dirKind
	}{
		{"flowchart TD", dirTB},
		{"flowchart tb", dirTB},
		{"graph LR", dirLR},
		{"flowchart BT", dirBT},
		{"graph rl", dirRL},
	}
	for _, tc := range cases {
		g := mustParse(t, tc.src)
		if g.dir != tc.want {
			t.Fatalf("%q: dir %v, want %v", tc.src, g.dir, tc.want)
		}
	}
}

func TestParseHeaderRejects(t *testing.T) {
	for _, src := range []string{
		"",
		"\n\n",
		"sequenceDiagram\nA->>B: hi",
		"flowchart\nA --> B",
		"flowchart XX\nA --> B",
		"graph",
	} {
		if _, err := parseFlowchart([]byte(src)); !errors.Is(err, errUnsupported) {
			t.Fatalf("%q: err %v, want errUnsupported", src, err)
		}
	}
}

func TestParseNodeShapes(t *testing.T) {
	g := mustParse(t, "flowchart TD\nA[Box]\nB(Round)\nC([Stad])\nD{Diamond}\nE((Dot))\nF")
	want := []struct {
		id, label string
		shape     shapeKind
	}{
		{"A", "Box", shapeRect},
		{"B", "Round", shapeRounded},
		{"C", "Stad", shapeStadium},
		{"D", "Diamond", shapeDiamond},
		{"E", "Dot", shapeCircle},
		{"F", "F", shapeRect}, // bare id: label defaults to id
	}
	if len(g.nodes) != len(want) {
		t.Fatalf("nodes %d, want %d", len(g.nodes), len(want))
	}
	for i, w := range want {
		n := g.nodes[i]
		if n.id != w.id || n.label != w.label || n.shape != w.shape {
			t.Fatalf("node %d = %+v, want %+v", i, n, w)
		}
	}
	if len(g.edges) != 0 {
		t.Fatalf("edges %d, want 0", len(g.edges))
	}
}

func TestParseQuotedLabel(t *testing.T) {
	g := mustParse(t, `flowchart TD
A["Label with ) parens"]`)
	if g.nodes[0].label != "Label with ) parens" {
		t.Fatalf("label %q", g.nodes[0].label)
	}
}

func TestParseNodeReferenceKeepsDefinition(t *testing.T) {
	// First definition wins; a later bare reference keeps label and shape.
	g := mustParse(t, "flowchart TD\nA[Box]\nA")
	if g.nodes[0].label != "Box" || g.nodes[0].shape != shapeRect {
		t.Fatalf("node = %+v", g.nodes[0])
	}
	if len(g.nodes) != 1 {
		t.Fatalf("duplicate node created: %d", len(g.nodes))
	}
}

func TestParseCommentsAndSemis(t *testing.T) {
	g := mustParse(t, "flowchart TD\n%% a comment\nA[Box]; B(Round) %% trailing\n%% another")
	if len(g.nodes) != 2 {
		t.Fatalf("nodes %d, want 2", len(g.nodes))
	}
}
