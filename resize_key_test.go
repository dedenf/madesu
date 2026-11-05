package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCtrlArrowResizes(t *testing.T) {
	m := sized(t, 120, 30)
	start := m.treeW
	m = send(m, tea.KeyMsg{Type: tea.KeyCtrlShiftRight})
	if m.treeW != start+resizeStep {
		t.Fatalf("ctrl+shift+right: treeW=%d, want %d", m.treeW, start+resizeStep)
	}
	if got, want := m.editor.Width(), innerWidth(120-m.treeW-gutterWidth); got != want {
		t.Fatalf("editor width %d, want %d", got, want)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyCtrlShiftLeft})
	m = send(m, tea.KeyMsg{Type: tea.KeyCtrlShiftLeft})
	if m.treeW != start-resizeStep {
		t.Fatalf("ctrl+shift+left x2: treeW=%d, want %d", m.treeW, start-resizeStep)
	}
}

func TestCtrlArrowRespectsLimits(t *testing.T) {
	m := sized(t, 120, 30)
	for i := 0; i < 100; i++ {
		m = send(m, tea.KeyMsg{Type: tea.KeyCtrlShiftLeft})
	}
	if m.treeW != minTreeWidth {
		t.Fatalf("batas bawah: treeW=%d, want %d", m.treeW, minTreeWidth)
	}
	for i := 0; i < 100; i++ {
		m = send(m, tea.KeyMsg{Type: tea.KeyCtrlShiftRight})
	}
	if max := 120 - gutterWidth - minRightWidth; m.treeW != max {
		t.Fatalf("batas atas: treeW=%d, want %d", m.treeW, max)
	}
}
