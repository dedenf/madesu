package main

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyStr(s string) tea.KeyMsg {
	switch s {
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "ctrl+home":
		return tea.KeyMsg{Type: tea.KeyCtrlHome}
	case "ctrl+end":
		return tea.KeyMsg{Type: tea.KeyCtrlEnd}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func longDoc(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		b.WriteString("baris ")
		b.WriteString(strings.Repeat("x", i%7))
		b.WriteString("\n")
	}
	return b.String()
}

func newSized(t *testing.T) model {
	t.Helper()
	m, err := newModel(t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	mm := tea.Model(m)
	mm, _ = mm.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	return mm.(model)
}

func TestEditorDocEdgesAndPage(t *testing.T) {
	m := newSized(t)
	m.openPath = "x.md"
	m.editor.SetValue(longDoc(200))
	m.focus = focusRight
	m.editor.Focus()

	upd := func(mm model, k string) model {
		r, _ := mm.Update(keyStr(k))
		return r.(model)
	}
	m = upd(m, "ctrl+end")
	if m.editor.Line() != m.editor.LineCount()-1 {
		t.Fatalf("ctrl+end: line %d, want %d", m.editor.Line(), m.editor.LineCount()-1)
	}
	m = upd(m, "ctrl+home")
	if m.editor.Line() != 0 {
		t.Fatalf("ctrl+home: line %d, want 0", m.editor.Line())
	}
	m = upd(m, "pgdown")
	if got := m.editor.Line(); got < 5 {
		t.Fatalf("pgdown tidak bergerak: line %d", got)
	}
	before := m.editor.Line()
	m = upd(m, "pgup")
	if m.editor.Line() >= before {
		t.Fatalf("pgup tidak naik: %d -> %d", before, m.editor.Line())
	}
}

func TestPreviewHomeEnd(t *testing.T) {
	m := newSized(t)
	m.openPath = "x.md"
	m.editor.SetValue(longDoc(200))
	m.focus = focusRight
	m.togglePreview()
	upd := func(mm model, k string) model {
		r, _ := mm.Update(keyStr(k))
		return r.(model)
	}
	m = upd(m, "end")
	if m.preview.AtTop() || !m.preview.AtBottom() {
		t.Fatalf("end: harus di bawah (top=%v bottom=%v)", m.preview.AtTop(), m.preview.AtBottom())
	}
	m = upd(m, "home")
	if !m.preview.AtTop() {
		t.Fatalf("home: harus di atas, YOffset=%d", m.preview.YOffset)
	}
	m = upd(m, "pgdown")
	if m.preview.AtTop() {
		t.Fatalf("pgdown tidak menggulir preview")
	}
	m = upd(m, "pgup")
	if !m.preview.AtTop() {
		t.Fatalf("pgup tidak kembali ke atas, YOffset=%d", m.preview.YOffset)
	}
}

func TestOpenFileCursorAtTop(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/a.md"
	if err := os.WriteFile(p, []byte(longDoc(200)), 0o644); err != nil {
		t.Fatal(err)
	}
	m := newSized(t)
	m.openFile(p)
	if got := m.editor.Line(); got != 0 {
		t.Fatalf("setelah dibuka kursor di baris %d, want 0", got)
	}
	r, _ := m.Update(keyStr("down"))
	m = r.(model)
	if got := m.editor.Line(); got != 1 {
		t.Fatalf("setelah down kursor di baris %d, want 1", got)
	}
}
