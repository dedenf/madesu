package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// folder berisi n file bernama f000.md, f001.md, ... untuk menguji scroll explorer.
func treeDir(t *testing.T, n int) string {
	t.Helper()
	dir := t.TempDir()
	for i := 0; i < n; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("f%03d.md", i)), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func newTreeModel(t *testing.T, dir string) model {
	t.Helper()
	m, err := newModel(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	return send(m, tea.WindowSizeMsg{Width: 120, Height: 20})
}

func TestDotfilesListedByDefault(t *testing.T) {
	dir := treeDir(t, 2)
	if err := os.WriteFile(filepath.Join(dir, ".env"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	m := newTreeModel(t, dir)
	found := false
	for _, e := range m.entries {
		if e.name == ".env" {
			found = true
		}
	}
	if !found {
		t.Fatal(".env harus tampil di explorer")
	}
}

func TestOpenHiddenFileForEditing(t *testing.T) {
	dir := treeDir(t, 0)
	p := filepath.Join(dir, ".env")
	if err := os.WriteFile(p, []byte("KEY=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := newTreeModel(t, dir)
	m.loadDir(dir)
	m.cursor = 1 // ".." di 0
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.openPath != p {
		t.Fatalf("file tersembunyi tidak terbuka: openPath=%q", m.openPath)
	}
	m.editor.SetValue("KEY=2\n")
	m = send(m, tea.KeyMsg{Type: tea.KeyCtrlS})
	data, _ := os.ReadFile(p)
	if string(data) != "KEY=2\n" {
		t.Fatalf("simpan dotfile gagal: %q", data)
	}
}

func TestWheelScrollsTreeView(t *testing.T) {
	dir := treeDir(t, 60)
	m := newTreeModel(t, dir)
	rows := m.treeRows()
	m = send(m, wheel(tea.MouseButtonWheelDown, 3))
	if m.treeTop != wheelLines {
		t.Fatalf("wheel down: treeTop=%d, want %d", m.treeTop, wheelLines)
	}
	if m.cursor < m.treeTop || m.cursor >= m.treeTop+rows {
		t.Fatalf("kursor %d keluar dari view [%d,%d)", m.cursor, m.treeTop, m.treeTop+rows)
	}
	m = send(m, wheel(tea.MouseButtonWheelUp, 3))
	if m.treeTop != 0 {
		t.Fatalf("wheel up: treeTop=%d, want 0", m.treeTop)
	}
}

func TestWheelTreeClampsAtEnds(t *testing.T) {
	dir := treeDir(t, 60)
	m := newTreeModel(t, dir)
	for i := 0; i < 100; i++ {
		m = send(m, wheel(tea.MouseButtonWheelDown, 3))
	}
	if max := len(m.entries) - m.treeRows(); m.treeTop != max {
		t.Fatalf("batas bawah: treeTop=%d, want %d", m.treeTop, max)
	}
	if m.cursor < m.treeTop || m.cursor >= len(m.entries) {
		t.Fatalf("kursor %d di luar daftar", m.cursor)
	}
	for i := 0; i < 100; i++ {
		m = send(m, wheel(tea.MouseButtonWheelUp, 3))
	}
	if m.treeTop != 0 {
		t.Fatalf("batas atas: treeTop=%d", m.treeTop)
	}
}

func TestKeyboardNavKeepsCursorInView(t *testing.T) {
	dir := treeDir(t, 60)
	m := newTreeModel(t, dir)
	for i := 0; i < 40; i++ {
		m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	if m.cursor < m.treeTop || m.cursor >= m.treeTop+m.treeRows() {
		t.Fatalf("setelah 40x j, kursor %d di luar view [%d,%d)", m.cursor, m.treeTop, m.treeTop+m.treeRows())
	}
}
