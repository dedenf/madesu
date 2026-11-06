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
