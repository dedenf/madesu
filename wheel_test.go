package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func wheel(b tea.MouseButton, x int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: 5, Action: tea.MouseActionPress, Button: b}
}

func TestWheelScrollsPreview(t *testing.T) {
	m := sized(t, 120, 30)
	m.editor.SetValue(paraDoc(200))
	m.openPath = "x.md"
	m.togglePreview()
	m = send(m, wheel(tea.MouseButtonWheelDown, m.treeW+5))
	if m.preview.YOffset == 0 {
		t.Fatal("wheel down tidak menggulir preview")
	}
	before := m.preview.YOffset
	m = send(m, wheel(tea.MouseButtonWheelUp, m.treeW+5))
	if m.preview.YOffset >= before {
		t.Fatalf("wheel up tidak naik: %d -> %d", before, m.preview.YOffset)
	}
}

func TestWheelMovesEditorView(t *testing.T) {
	m := sized(t, 120, 30)
	m.editor.SetValue(longDoc(200))
	m.openPath = "x.md"
	moveToDocEdge(&m.editor, -1)
	m.focus = focusRight
	m.editor.Focus()
	before := plain(m.editor.View())
	m = send(m, wheel(tea.MouseButtonWheelDown, m.treeW+5))
	if got := m.editor.Line(); got != wheelLines {
		t.Fatalf("wheel down: kursor baris %d, want %d", got, wheelLines)
	}
	after := plain(m.editor.View())
	if before == after {
		t.Fatal("wheel down tidak menggulir view editor")
	}
	// baris 0 = "baris x", baris 3 = "baris xxxx": view harus mulai dari baris 3.
	firstLine := strings.SplitN(after, "\n", 2)[0]
	if !strings.Contains(firstLine, "baris xxxx") {
		t.Fatalf("view editor belum mulai dari baris 3: %q", firstLine)
	}
}

func TestWheelOnTreeIgnored(t *testing.T) {
	m := sized(t, 120, 30)
	m.editor.SetValue(longDoc(200))
	m.openPath = "x.md"
	m.showPrev = true
	m = send(m, wheel(tea.MouseButtonWheelDown, 3))
	if m.preview.YOffset != 0 {
		t.Fatalf("wheel di panel kiri tidak boleh menggulir preview")
	}
}

// paraDoc: paragraf dipisah baris kosong supaya markdown tidak menggabungkannya.
func paraDoc(n int) string {
	out := ""
	for i := 1; i <= n; i++ {
		out += fmt.Sprintf("Paragraf nomor %d.\n\n", i)
	}
	return out
}
