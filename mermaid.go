package main

// Parser for the mermaid flowchart subset (see
// docs/superpowers/specs/2026-11-06-mermaid-flowchart-design.md). Anything
// outside the subset returns errUnsupported so the preview falls back to a
// plain code block.

import (
	"errors"
	"strings"
)

var errUnsupported = errors.New("mermaid: unsupported syntax")

type shapeKind int

const (
	shapeRect shapeKind = iota
	shapeRounded
	shapeStadium
	shapeDiamond
	shapeCircle
)

type edgeStyle int

const (
	edgeNormal edgeStyle = iota
	edgeDotted
	edgeThick
)

type dirKind int

const (
	dirTB dirKind = iota
	dirLR
	dirBT
	dirRL
)

type flowNode struct {
	id    string
	label string
	shape shapeKind
}

type flowEdge struct {
	from, to string
	label    string
	style    edgeStyle
	arrow    bool
}

type flowGraph struct {
	dir   dirKind
	nodes []*flowNode // first-seen order
	edges []flowEdge
}

func parseFlowchart(src []byte) (*flowGraph, error) {
	lines := strings.Split(string(src), "\n")
	head := -1
	for i, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "%%") {
			continue
		}
		head = i
		break
	}
	if head < 0 {
		return nil, errUnsupported
	}
	f := strings.Fields(lines[head])
	if len(f) != 2 {
		return nil, errUnsupported
	}
	switch strings.ToLower(f[0]) {
	case "flowchart", "graph":
	default:
		return nil, errUnsupported
	}
	var dir dirKind
	switch strings.ToLower(f[1]) {
	case "td", "tb":
		dir = dirTB
	case "lr":
		dir = dirLR
	case "bt":
		dir = dirBT
	case "rl":
		dir = dirRL
	default:
		return nil, errUnsupported
	}
	p := &parser{g: &flowGraph{dir: dir}, byID: map[string]*flowNode{}}
	for _, l := range lines[head+1:] {
		for _, stmt := range strings.Split(l, ";") {
			stmt = strings.TrimSpace(stripComment(stmt))
			if stmt == "" {
				continue
			}
			if err := p.statement(stmt); err != nil {
				return nil, err
			}
		}
	}
	return p.g, nil
}

type parser struct {
	g    *flowGraph
	byID map[string]*flowNode
}

// statement parses one `;`/newline-separated statement. Body statements are
// added by later tasks; until then everything is unsupported.
func (p *parser) statement(s string) error {
	return errUnsupported
}

// stripComment cuts a `%%` comment that starts outside double quotes.
func stripComment(s string) string {
	inQ := false
	for i := 0; i+1 < len(s); i++ {
		switch s[i] {
		case '"':
			inQ = !inQ
		case '%':
			if !inQ && s[i+1] == '%' {
				return s[:i]
			}
		}
	}
	return s
}
