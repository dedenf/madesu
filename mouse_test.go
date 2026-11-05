package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func mouse(action tea.MouseAction, x int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: 5, Action: action, Button: tea.MouseButtonLeft}
}

func sized(t *testing.T, w, h int) model {
	t.Helper()
	m, err := newModel(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	var tm tea.Model = m
	tm, _ = tm.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return tm.(model)
}

func send(m model, msg tea.Msg) model {
	r, _ := m.Update(msg)
	return r.(model)
}

func TestDragDivider(t *testing.T) {
	m := sized(t, 120, 40)
	if m.treeW != defaultTreeWidth {
		t.Fatalf("lebar awal %d, want %d", m.treeW, defaultTreeWidth)
	}
	m = send(m, mouse(tea.MouseActionPress, m.treeW))
	if !m.dragging {
		t.Fatal("tekan di garis pemisah harus mulai drag")
	}
	m = send(m, mouse(tea.MouseActionMotion, 50))
	if m.treeW != 50 {
		t.Fatalf("setelah geser ke 50, treeW=%d", m.treeW)
	}
	if got, want := m.editor.Width(), innerWidth(120-50-gutterWidth); got != want {
		t.Fatalf("lebar editor %d, want %d (ikut berubah)", got, want)
	}
	m = send(m, mouse(tea.MouseActionRelease, 50))
	if m.dragging {
		t.Fatal("lepas tombol harus mengakhiri drag")
	}
	m = send(m, mouse(tea.MouseActionMotion, 80))
	if m.treeW != 50 {
		t.Fatalf("motion setelah lepas tidak boleh mengubah lebar, treeW=%d", m.treeW)
	}
}

func TestDragClamps(t *testing.T) {
	m := sized(t, 120, 40)
	m = send(m, mouse(tea.MouseActionPress, m.treeW))
	m = send(m, mouse(tea.MouseActionMotion, 2))
	if m.treeW != minTreeWidth {
		t.Fatalf("batas bawah: treeW=%d, want %d", m.treeW, minTreeWidth)
	}
	m = send(m, mouse(tea.MouseActionMotion, 500))
	if max := 120 - gutterWidth - minRightWidth; m.treeW != max {
		t.Fatalf("batas atas: treeW=%d, want %d", m.treeW, max)
	}
}

func TestPressAwayFromDividerIgnored(t *testing.T) {
	m := sized(t, 120, 40)
	m = send(m, mouse(tea.MouseActionPress, 5))
	if m.dragging {
		t.Fatal("tekan jauh dari garis pemisah tidak boleh mulai drag")
	}
}
